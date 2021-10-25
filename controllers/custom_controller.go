/*
Copyright 2021.

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

package controllers

import (
	"context"
	batchv1 "github.com/Rakibul-Hossain/api/v1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CustomReconciler reconciles a Custom object
type CustomReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=batch.custom.com,resources=customs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.custom.com,resources=customs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.custom.com,resources=customs/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=customs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=customs/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Custom object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile

var (
	ownerKey  = ".metadata.controller"
	ownerKind = "Custom"
	apiGVStr  = batchv1.GroupVersion.String()
)

func (r *CustomReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("custom", req.NamespacedName)

	// load the Custom object
	var custom batchv1.Custom
	if err := r.Get(ctx, req.NamespacedName, &custom); err != nil {
		log.Error(err, "unable to fetch Custom")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// list the deployments
	var deployments appsv1.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(req.Namespace), client.MatchingFields{ownerKey: req.Name}); err != nil {
		log.Error(err, "unable to list deployments")
		return ctrl.Result{}, err
	}

	// if not found deployments
	if len(deployments.Items) == 0 {
		log.Info("Creating the custom deployment")
		if err := r.Create(ctx, newDeployment(&custom)); err != nil {
			log.Error(err, "unable to create deployment")
			return ctrl.Result{}, err
		}
	}

	// list the service
	var services corev1.ServiceList
	if err := r.List(ctx, &services, client.InNamespace(req.Namespace), client.MatchingFields{ownerKey: req.Name}); err != nil {
		log.Error(err, "unable to list services")
		return ctrl.Result{}, err
	}

	// if not found services
	if len(services.Items) == 0 {
		log.Info("Creating the custom service")
		if err := r.Create(ctx, newService(&custom)); err != nil {
			log.Error(err, "unable to create service")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CustomReconciler) SetupWithManager(mgr ctrl.Manager) error {

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, ownerKey, func(rawObj client.Object) []string {
		// grab the deploy object, extract the owner...
		deployment := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(deployment)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Custom deployment...
		if owner.APIVersion != apiGVStr || owner.Kind != ownerKind {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Service{}, ownerKey, func(rawObj client.Object) []string {
		// grab the deploy object, extract the owner...
		service := rawObj.(*corev1.Service)
		owner := metav1.GetControllerOf(service)
		if owner == nil {
			return nil
		}
		// ...make sure it's a Custom deployment...
		if owner.APIVersion != apiGVStr || owner.Kind != ownerKind {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Custom{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
