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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TLQMaster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *TLQMasterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("tlqmaster", req.NamespacedName)

	master := &tlqv1alpha1.TLQMaster{}
	err := r.Get(ctx, req.NamespacedName, master)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("TongLinkQMaster resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get TongLinkQMaster")
		return ctrl.Result{}, err
	}
	pod := &v1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: master.Name, Namespace: master.Namespace}, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			pod := buildMasterPod(master)
			if err := r.Create(ctx, pod); err != nil && !errors.IsAlreadyExists(err) {
				return ctrl.Result{}, err
			}
			if err := controllerutil.SetControllerReference(master, pod, r.Scheme); err != nil {
				return ctrl.Result{}, err
			}
			master.Status.Parse = tlqv1alpha1.Pending
			err := r.Status().Update(ctx, master)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, err
		}
	} else {
		oldStatus := master.Status.Parse
		if v1.PodRunning == pod.Status.Phase {
			master.Status.Parse = tlqv1alpha1.UnHealthy
			master.Status.MasterIp = pod.Status.PodIP
			master.Status.NodeIp = pod.Status.HostIP
		} else if v1.PodFailed == pod.Status.Phase {
			master.Status.Parse = tlqv1alpha1.Fail
		} else {
			master.Status.Parse = tlqv1alpha1.Pending
		}
		if oldStatus != master.Status.Parse {
			err := r.Status().Update(ctx, master)
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}
func buildMasterPod(master *tlqv1alpha1.TLQMaster) *v1.Pod {
	containers := make([]v1.Container, 1)
	ports := make([]v1.ContainerPort, 1)
	ports[0] = v1.ContainerPort{
		ContainerPort: master.Spec.Port,
	}
	policy := master.Spec.ImagePullPolicy
	if "" == policy {
		policy = v1.PullAlways
	}
	containers[0] = v1.Container{
		Image:           master.Spec.Image,
		ImagePullPolicy: policy,
		VolumeMounts:    master.Spec.VolumeMounts,
		Ports:           ports,
		Env:             master.Spec.Envs,
	}
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      master.Name,
			Namespace: master.Namespace,
			Labels: map[string]string{
				"role": "master",
			},
		},
		Spec: v1.PodSpec{
			Volumes:       master.Spec.Volumes,
			Containers:    containers,
			RestartPolicy: v1.RestartPolicyAlways,
		},
	}
	return &pod
}

// SetupWithManager sets up the controller with the Manager.
func (r *TLQMasterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c := ctrl.NewControllerManagedBy(mgr)
	c.Watches(&source.Kind{Type: &v1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &tlqv1alpha1.TLQMaster{},
	})
	return c.For(&tlqv1alpha1.TLQMaster{}).
		Complete(r)
}
