/*
Copyright 2024 Vikas Badoni.

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
	"time"

	apiv1alpha1 "github.com/Projects-vb/Kube-operator/scaling/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var logger = log.Log.WithName("AppscaleController")

// AppscaleReconciler reconciles a Appscale object
type AppscaleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=api.vikasbadoni.com,resources=appscales,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.vikasbadoni.com,resources=appscales/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.vikasbadoni.com,resources=appscales/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Appscale object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *AppscaleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	log := logger.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	log.Info("Reconcilation logic")

	s := &apiv1alpha1.Appscale{}

	err := r.Get(ctx, req.NamespacedName, s)
	if err != nil {
		log.Error(err, "Failed")
		return ctrl.Result{}, err
	}

	startTime := s.Spec.Start
	endTime := s.Spec.End
	replicas := s.Spec.Replicas
	currentTime := time.Now().UTC().Hour()

	if currentTime >= startTime && currentTime <= endTime {
		scaleApp(s, r, ctx, req, int32(replicas))
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Duration(30 * time.Second)}, nil
}

func scaleApp(scale *apiv1alpha1.Appscale, r *AppscaleReconciler, ctx context.Context, req ctrl.Request, replicas int32) error {
	for _, deployment := range scale.Spec.Deployments {
		dep := &v1.Deployment{}
		r.Get(ctx, types.NamespacedName{
			Namespace: deployment.Namespace,
			Name:      deployment.Name,
		}, dep)

		if dep.Spec.Replicas != &replicas {
			dep.Spec.Replicas = &replicas
			err := r.Update(ctx, dep)
			if err != nil {
				fmt.Println("Failde to update the resource")
				return err
			}
			r.Status().Update(ctx, scale)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppscaleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1alpha1.Appscale{}).
		Complete(r)
}
