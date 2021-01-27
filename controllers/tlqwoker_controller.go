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
	"github.com/go-logr/logr"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	tlqv1alpha1 "tlq9-operator/api/v1alpha1"
)

// TLQWorkerReconciler reconciles a TLQWorker object
type TLQWorkerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqworkers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqworkers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqworkers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the TLQWorker object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *TLQWorkerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("tlqworker", req.NamespacedName)
	//do something
	operate := &WorkerOperate{
		log: log,
		r:   r,
		ctx: ctx,
		req: req,
	}
	//get TLQWorker Resource
	worker, result, err := operate.GetWorker()
	if worker == nil {
		return result, err
	}
	//service
	svc, svcResult, err := operate.CreateOrUpdateService(worker)
	service := &v1.Service{}
	if svc == nil {
		return svcResult, err
	} else {
		err := r.Get(ctx, types.NamespacedName{Name: worker.Name, Namespace: worker.Namespace}, service)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	//get reference statefulSet
	stateful, c, err := operate.CreateOrUpdateStatefulSet(worker, service)
	if stateful == nil {
		return c, err
	}
	//update  TLQMaster status
	return operate.UpdateWorkerStatus(worker, stateful, service)
}

// SetupWithManager sets up the controller with the Manager.
func (r *TLQWorkerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tlqv1alpha1.TLQWorker{}).
		Owns(&v12.StatefulSet{}).
		Complete(r)
}
