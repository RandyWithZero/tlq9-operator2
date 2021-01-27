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
	"errors"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"

	tlqv1alpha1 "tlq9-operator/api/v1alpha1"
)

// TLQClusterReconciler reconciles a TLQCluster object
type TLQClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqmasters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tlq.tongtech.com,resources=tlqworkers,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the TLQCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *TLQClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("tlqcluster", req.NamespacedName)
	//do something
	operate := &ClusterOperate{
		log: log,
		r:   r,
		ctx: ctx,
		req: req,
	}
	//get tlqCluster Resource
	cluster, result, err := operate.GetCluster()
	if cluster == nil {
		return result, err
	}
	//master operate
	master, masterResult, err := operate.CreateOrUpdateTlqMaster(cluster)
	if master == nil {
		return masterResult, err
	}
	//update cluster status from master
	status, err := operate.updateClusterStatus(cluster, master, nil)
	if err != nil {
		return status, err
	}
	if master.Status.Parse != tlqv1alpha1.Healthy {
		return ctrl.Result{}, errors.New("nameserver not ready")
	}
	nameserverUrl := master.Name + "/" + strconv.Itoa(int(master.Spec.Detail.Port))
	//worker operate
	workerList, c, err := operate.ListWorker(cluster)
	if workerList == nil {
		return c, err
	}
	//update cluster status from worker
	clusterStatus, err := operate.updateClusterStatus(cluster, nil, workerList)
	if err != nil {
		return clusterStatus, err
	}
	size := cluster.Spec.WorkerSize
	if len(workerList.Items) < size {
		indexList := make([]int, size)
		for _, item := range workerList.Items {
			index, _ := strconv.Atoi(item.Labels["index"])
			indexList[index] = 1
		}
		willCreateIndex := size
		for i := range indexList {
			if indexList[i] == 0 {
				willCreateIndex = i
				break
			}
		}
		if willCreateIndex < size {
			worker, workerResult, err := operate.CreateOrUpdateTlqWorker(cluster, willCreateIndex, nameserverUrl)
			if worker == nil {
				return workerResult, err
			}
			//return ctrl.Result{}, errors.New("workers are  created to respect count:" + strconv.Itoa(size))
		}

	} else if len(workerList.Items) > size {
		for _, item := range workerList.Items {
			index, _ := strconv.Atoi(item.Labels["index"])
			if index > cluster.Spec.WorkerSize {
				c, err := operate.DeleteWorker(&item)
				if err != nil {
					return c, err
				}
				//return ctrl.Result{}, errors.New("workers are  deleted to respect count:" + strconv.Itoa(size))
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TLQClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tlqv1alpha1.TLQCluster{}).
		Owns(&tlqv1alpha1.TLQWorker{}).
		Owns(&tlqv1alpha1.TLQMaster{}).
		Complete(r)
}
