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
	v12 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tlqv1alpha1 "tlq9-operator/api/v1alpha1"
)

// TLQMasterReconciler reconciles a TLQMaster object
type TLQMasterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqmasters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqmasters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqmasters/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TLQMaster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
// hhlo
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *TLQMasterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("tlqmaster", req.NamespacedName)

	//do something
	operate := &MasterOperate{
		log: log,
		r:   r,
		ctx: ctx,
		req: req,
	}
	//get TLQMaster Resource
	master, result, err := operate.GetMaster()
	if master == nil {
		return result, err
	}
	//get reference statefulSet
	stateful, c, err := operate.CreateOrUpdateStatefulSet(master)
	if stateful == nil {
		return c, err
	}
	//update  TLQMaster status
	re, err := operate.UpdateMasterStatus(master, stateful)
	return re, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *TLQMasterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Watches(&source.Kind{Type: &v12.StatefulSet{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &tlqv1alpha1.TLQMaster{},
		}).For(&tlqv1alpha1.TLQMaster{}).
		Complete(r)
}
