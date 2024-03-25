/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	avyriov1 "github.com/avyr-io/epha/api/v1"
)

// AnnotatedObjectReconciler reconciles a AnnotatedObject object
type AnnotatedObjectReconciler struct {
	client.Client
	DynamicClient dynamic.Interface
	Scheme        *runtime.Scheme
}

//+kubebuilder:rbac:groups=epha.avyr.io,resources=annotatedobjects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=epha.avyr.io,resources=annotatedobjects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=epha.avyr.io,resources=annotatedobjects/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments;replicasets,verbs=get;list;watch;create;update;patch;delete

func (r *AnnotatedObjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var annotatedObject avyriov1.AnnotatedObject
	if err := r.Get(ctx, req.NamespacedName, &annotatedObject); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Example: Process each target specified in the AnnotatedObject
	for _, target := range annotatedObject.Spec.Targets {
		if err := r.reconcileTargetDynamic(ctx, target); err != nil {
			return reconcile.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *AnnotatedObjectReconciler) reconcileTargetDynamic(ctx context.Context, target avyriov1.TargetResourceWithMetadata) error {
	// Split the APIVersion to group and version
	parts := strings.SplitN(target.APIVersion, "/", 2)
	group := ""
	version := parts[0]
	if len(parts) == 2 {
		group = parts[0]
		version = parts[1]
	}

	// Construct the GVR
	gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: strings.ToLower(target.Kind) + "s"}

	// Use dynamic client
	dynClient, err := dynamic.NewForConfig(config.GetConfigOrDie())
	if err != nil {
		return err
	}

	// Fetch the target resource
	resource, err := dynClient.Resource(gvr).Namespace(target.Namespace).Get(ctx, target.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Update annotations on the unstructured object
	annotations := target.Metadata.Annotations
	unstructuredAnnotations := resource.GetAnnotations()
	if unstructuredAnnotations == nil {
		unstructuredAnnotations = make(map[string]string)
	}
	for key, value := range annotations {
		unstructuredAnnotations[key] = value
	}
	resource.SetAnnotations(unstructuredAnnotations)

	// Update the resource with the new annotations
	_, err = dynClient.Resource(gvr).Namespace(target.Namespace).Update(ctx, resource, metav1.UpdateOptions{})
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *AnnotatedObjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&avyriov1.AnnotatedObject{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.ReplicaSet{}).
		Complete(r)
}
