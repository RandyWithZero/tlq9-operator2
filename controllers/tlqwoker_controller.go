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
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	tlqv1alpha1 "tlq9-operator/api/v1alpha1"
)

var _ handler.EventHandler = &TLQEventHandler{}

type TLQEventHandler struct {
}

func (e *TLQEventHandler) Create(event.CreateEvent, workqueue.RateLimitingInterface) {
	fmt.Println("Create-------------------------------------------")
}

func (e *TLQEventHandler) Update(event.UpdateEvent, workqueue.RateLimitingInterface) {
	fmt.Println("Update-------------------------------------------")
}
func (e *TLQEventHandler) Delete(event.DeleteEvent, workqueue.RateLimitingInterface) {
	fmt.Println("Delete-------------------------------------------")
}

// Generic is called in response to an event of an unknown type or a synthetic event triggered as a cron or
// external trigger request - e.g. reconcile Autoscaling, or a Webhook.
func (e *TLQEventHandler) Generic(event.GenericEvent, workqueue.RateLimitingInterface) {
	fmt.Println("Generic-------------------------------------------")
}

// TLQWokerReconciler reconciles a TLQWoker object
type TLQWokerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqwokers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqwokers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqwokers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TLQWoker object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *TLQWokerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := r.Log.WithValues("tlqwoker", req.NamespacedName)
	log.Info("TLQWokerReconciler=========================================================")
	// your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TLQWokerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tlqv1alpha1.TLQWoker{}).
		Complete(r)
}
