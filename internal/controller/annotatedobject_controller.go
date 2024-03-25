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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	avyriov1 "github.com/avyr-io/epha/api/v1"
)

// AnnotatedObjectReconciler reconciles a AnnotatedObject object
type AnnotatedObjectReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=epha.avyr.io,resources=annotatedobjects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=epha.avyr.io,resources=annotatedobjects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=epha.avyr.io,resources=annotatedobjects/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AnnotatedObject object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile

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
		if err := r.reconcileTarget(ctx, target); err != nil {
			return reconcile.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *AnnotatedObjectReconciler) reconcileTarget(ctx context.Context, target avyriov1.TargetResourceWithMetadata) error {
	var err error

	switch target.Kind {
	case "Deployment":
		var deployment appsv1.Deployment
		err = r.Get(ctx, types.NamespacedName{Name: target.Name, Namespace: target.Namespace}, &deployment)
		if err != nil {
			return err
		}
		mergeAnnotations(&deployment.ObjectMeta, target.Metadata.Annotations)
		err = r.Update(ctx, &deployment)
	case "ReplicaSet":
		var replicaSet appsv1.ReplicaSet
		err = r.Get(ctx, types.NamespacedName{Name: target.Name, Namespace: target.Namespace}, &replicaSet)
		if err != nil {
			return err
		}
		mergeAnnotations(&replicaSet.ObjectMeta, target.Metadata.Annotations)
		err = r.Update(ctx, &replicaSet)
	default:
		return fmt.Errorf("unsupported kind: %s", target.Kind)
	}

	return err
}

func mergeAnnotations(meta *metav1.ObjectMeta, newAnnotations map[string]string) {
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	for key, value := range newAnnotations {
		meta.Annotations[key] = value
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AnnotatedObjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&avyriov1.AnnotatedObject{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.ReplicaSet{}).
		Complete(r)
}
