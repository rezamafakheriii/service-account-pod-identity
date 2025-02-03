package lib

// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"context"
	"log"
	"reflect"

	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

// ServiceAccountController monitors service account definition changes in a namespace.
// For each service account object, its SpiffeID is added to identity registry for
// whitelisting purpose.
type ServiceAccountController struct {
	core corev1.CoreV1Interface

	// controller for service objects
	controller cache.Controller
}

// NewServiceAccountController returns a new ServiceAccountController
func NewServiceAccountController(core corev1.CoreV1Interface, namespace string) *ServiceAccountController {
	c := &ServiceAccountController{
		core: core,
	}

	LW := &cache.ListWatch{
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return core.ServiceAccounts(namespace).Watch(context.Background(), options)
		},
	}

	opts := cache.InformerOptions{
		ListerWatcher: LW,
		ObjectType:    &v1.ServiceAccount{},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    c.serviceAccountAdded,
			DeleteFunc: c.serviceAccountDeleted,
			UpdateFunc: c.serviceAccountUpdated,
		},
	}

	_, c.controller = cache.NewInformerWithOptions(opts)
	return c
}

// Run starts the ServiceAccountController until a value is sent to stopCh.
// It should only be called once.
func (c *ServiceAccountController) Run(stopCh chan struct{}) {
	go c.controller.Run(stopCh)
}

func (c *ServiceAccountController) serviceAccountAdded(obj interface{}) {
	sa := obj.(*v1.ServiceAccount)
	log.Printf("ServiceAccount added: %s\n", sa.Name)
}

func (c *ServiceAccountController) serviceAccountDeleted(obj interface{}) {
	sa := obj.(*v1.ServiceAccount)
	log.Printf("ServiceAccount deleted: %s\n", sa.Name)
}

func (c *ServiceAccountController) serviceAccountUpdated(oldObj, newObj interface{}) {
	if oldObj == newObj || reflect.DeepEqual(oldObj, newObj) {
		// Nothing is changed. The method is invoked by periodical re-sync with the apiserver.
		log.Printf("ServiceAccount updated: no change\n")
		return
	}

	oldSa := oldObj.(*v1.ServiceAccount)
	log.Printf("ServiceAccount updated old: %s\n", oldSa.Name)
	newSa := newObj.(*v1.ServiceAccount)

	log.Printf("ServiceAccount updated new: %s\n", newSa.Name)

}
