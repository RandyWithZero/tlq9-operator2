package controllers

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"
	"sync"
	"tlq9-operator/api/v1alpha1"
)

var (
	NameServerSign = "nameserver"
	//PrimarySign    = "primary"
	//SecondarySign  = "secondary"
	WorkerSign = "worker"
)
var rwRock sync.RWMutex

type ClusterOperate struct {
	log logr.Logger
	r   *TLQClusterReconciler
	ctx context.Context
	req ctrl.Request
}

func (o *ClusterOperate) GetCluster() (*v1alpha1.TLQCluster, ctrl.Result, error) {
	cluster := &v1alpha1.TLQCluster{}
	err := o.r.Get(o.ctx, o.req.NamespacedName, cluster)
	if err != nil {
		if errors.IsNotFound(err) {
			o.log.Info("TLQCluster resource not found. Ignoring since object must be deleted.")
			return nil, ctrl.Result{}, nil
		}
		o.log.Error(err, "Failed to get TLQCluster")
		return nil, ctrl.Result{}, err
	}
	return cluster, ctrl.Result{}, nil
}
func (o *ClusterOperate) ListWorker(cluster *v1alpha1.TLQCluster) (*v1alpha1.TLQWorkerList, ctrl.Result, error) {
	labelSelector := labels.SelectorFromSet(map[string]string{"cluster": cluster.Name})
	listOps := &client.ListOptions{
		Namespace:     cluster.Namespace,
		LabelSelector: labelSelector,
	}
	list := &v1alpha1.TLQWorkerList{}
	err := o.r.List(o.ctx, list, listOps)
	if err != nil {
		o.log.Error(err, "Failed to get TLQWorkerList")
		return nil, ctrl.Result{}, err
	}
	return list, ctrl.Result{}, nil
}
func (o *ClusterOperate) DeleteWorker(worker *v1alpha1.TLQWorker) (ctrl.Result, error) {
	err := o.r.Delete(o.ctx, worker)
	if err != nil {
		o.log.Error(err, "Failed to delete  TLQWorker")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (o *ClusterOperate) CreateOrUpdateTlqMaster(cluster *v1alpha1.TLQCluster) (*v1alpha1.TLQMaster, ctrl.Result, error) {
	masterName := cluster.Name + "-" + NameServerSign
	master := &v1alpha1.TLQMaster{}
	err := o.r.Get(o.ctx, types.NamespacedName{Name: masterName, Namespace: cluster.Namespace}, master)
	if err != nil {
		if errors.IsNotFound(err) {
			instance := buildNameServerInstance(cluster)
			if err := controllerutil.SetControllerReference(cluster, instance, o.r.Scheme); err != nil {
				return nil, ctrl.Result{}, err
			}
			o.log.Info("set nameserver owner ...")

			if err := o.r.Create(o.ctx, instance); err != nil && !errors.IsAlreadyExists(err) {
				return nil, ctrl.Result{}, err
			}
			o.log.Info("create nameserver ...")
		} else {
			return nil, ctrl.Result{}, err
		}
	} else {
		if !compareNameServerSpec(cluster.Spec.MasterTemplate, master.Spec) {
			instance := buildNameServerInstance(cluster)
			instance.ObjectMeta = *master.ObjectMeta.DeepCopy()
			err := o.r.Update(o.ctx, instance)
			o.log.Info("update nameserver...")
			if err != nil {
				return nil, ctrl.Result{}, err
			} else {
				return instance, ctrl.Result{}, nil
			}

		}
	}
	return master, ctrl.Result{}, nil
}

func (o *ClusterOperate) CreateOrUpdateTlqWorker(cluster *v1alpha1.TLQCluster, index int, nameserverUrl string) (*v1alpha1.TLQWorker, ctrl.Result, error) {
	workerName := cluster.Name + "-" + WorkerSign + "-" + strconv.Itoa(index)
	worker := &v1alpha1.TLQWorker{}
	err := o.r.Get(o.ctx, types.NamespacedName{Name: workerName, Namespace: cluster.Namespace}, worker)
	if err != nil {
		if errors.IsNotFound(err) {
			instance := buildWorkerInstance(cluster, index)
			instance.Annotations["nameserver"] = nameserverUrl
			if err := controllerutil.SetControllerReference(cluster, instance, o.r.Scheme); err != nil {
				return nil, ctrl.Result{}, err
			}
			o.log.Info("set worker owner ...")

			if err := o.r.Create(o.ctx, instance); err != nil && !errors.IsAlreadyExists(err) {
				return nil, ctrl.Result{}, err
			}
			o.log.Info("create worker ...")
		} else {
			return nil, ctrl.Result{}, err
		}
	} else {
		if !compareWorkerSpec(cluster.Spec.WorkerTemplate, worker.Spec) {
			instance := buildWorkerInstance(cluster, index)
			instance.ObjectMeta = *worker.ObjectMeta.DeepCopy()
			instance.Annotations["nameserver"] = nameserverUrl
			err := o.r.Update(o.ctx, instance)
			o.log.Info("update worker...")
			if err != nil {
				return nil, ctrl.Result{}, err
			} else {
				return instance, ctrl.Result{}, nil
			}

		}
	}
	return worker, ctrl.Result{}, nil
}

func (o *ClusterOperate) updateClusterStatus(cluster *v1alpha1.TLQCluster, master *v1alpha1.TLQMaster, workerList *v1alpha1.TLQWorkerList) (ctrl.Result, error) {
	rwRock.Lock()
	o.log.Info("update TLQCluster resource status ...")
	status := &cluster.Status
	status.Parse = v1alpha1.Pending
	if &status.ReadyWorkerServer == nil {
		status.ReadyWorkerServer = map[string]string{}
	}
	if master != nil {
		if master.Status.Parse == v1alpha1.Healthy {
			status.ReadyMasterServer = master.Status.Server
			status.Parse = v1alpha1.UnHealthy
		}
	}
	status.TotalWorkerCount = cluster.Spec.WorkerSize
	if workerList != nil {
		workers := status.ReadyWorkerServer
		for _, item := range workerList.Items {
			if item.Status.Parse == v1alpha1.Healthy {
				workers[item.Name] = item.Status.Server
			} else {
				delete(workers, item.Name)
			}
		}
		status.ReadyWorkerCount = len(workers)
	}
	if status.ReadyWorkerCount == status.TotalWorkerCount {
		status.Parse = v1alpha1.Healthy
	}
	rwRock.Unlock()
	return ctrl.Result{}, nil
}

func buildNameServerInstance(cluster *v1alpha1.TLQCluster) *v1alpha1.TLQMaster {
	masterName := cluster.Name + "-" + NameServerSign
	template := cluster.Spec.MasterTemplate
	master := &v1alpha1.TLQMaster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      masterName,
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"role":    NameServerSign,
				"cluster": cluster.Name,
			},
		},
		Spec: *template.DeepCopy(),
	}
	master.Spec.Detail = cluster.Spec.MasterTemplate.Detail.DeepCopy()
	return master

}
func buildWorkerInstance(cluster *v1alpha1.TLQCluster, index int) *v1alpha1.TLQWorker {
	workerName := cluster.Name + "-" + WorkerSign + "-" + strconv.Itoa(index)
	template := cluster.Spec.WorkerTemplate
	worker := &v1alpha1.TLQWorker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workerName,
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"role":    WorkerSign,
				"cluster": cluster.Name,
				"index":   strconv.Itoa(index),
			},
		},
		Spec: *template.DeepCopy(),
	}
	worker.Spec.Detail = cluster.Spec.MasterTemplate.Detail.DeepCopy()
	return worker
}
func compareNameServerSpec(spec1 v1alpha1.TLQMasterSpec, spec2 v1alpha1.TLQMasterSpec) bool {
	var1 := spec1.VRRPPasswd == spec2.VRRPPasswd
	var2 := spec1.UserName == spec2.UserName
	var3 := spec1.Password == spec2.Password
	var4 := spec1.AdvertiseInterval == spec2.AdvertiseInterval
	var5 := compareSpec(spec1.Detail, spec2.Detail)
	return var1 && var2 && var3 && var4 && var5
}
func compareWorkerSpec(spec1 v1alpha1.TLQWorkerSpec, spec2 v1alpha1.TLQWorkerSpec) bool {
	var1 := spec1.IsAffinity == spec2.IsAffinity
	var2 := spec1.LogLevel == spec2.LogLevel
	var3 := spec1.RegisterStatus == spec2.RegisterStatus
	var4 := spec1.RequestServiceNum == spec2.RequestServiceNum
	var5 := spec1.ResponseServiceNum == spec2.ResponseServiceNum
	var6 := spec1.WorkRootDir == spec2.WorkRootDir
	var7 := compareSpec(spec1.Detail, spec2.Detail)
	return var1 && var2 && var3 && var4 && var5 && var6 && var7
}
func compareSpec(spec1 *v1alpha1.Spec, spec2 *v1alpha1.Spec) bool {
	var1 := spec1.ZoneConfigMapName == spec2.ZoneConfigMapName
	var2 := spec1.TopicConfigMapName == spec2.TopicConfigMapName
	var3 := spec1.Port == spec2.Port
	var4 := spec1.Image == spec2.Image
	var5 := spec1.ImagePullPolicy == spec2.ImagePullPolicy
	var6 := false
	if spec1.DataPersistentSpec != nil && spec2.DataPersistentSpec != nil {
		var6 = compareDataPersistentSpec(*spec1.DataPersistentSpec, *spec2.DataPersistentSpec)
	} else if spec1.DataPersistentSpec == nil && spec2.DataPersistentSpec == nil {
		var6 = true
	}
	var7 := false
	if spec1.Resources != nil && spec2.Resources != nil {
		var7 = compareResources(*spec1.Resources, *spec2.Resources)
	} else if spec1.Resources == nil && spec2.Resources == nil {
		var7 = true
	}

	return var1 && var2 && var3 && var4 && var5 && var6 && var7
}
func compareDataPersistentSpec(spec1 v1alpha1.DataPersistentSpec, spec2 v1alpha1.DataPersistentSpec) bool {
	var1 := spec1.DataDir == spec2.DataDir
	var2 := spec1.DataMountType == spec2.DataMountType
	var3 := spec1.DataStorage.Equal(spec2.DataStorage)
	var4 := spec1.AccessMode == spec2.AccessMode
	var5 := spec1.HostPath == spec2.HostPath
	return var1 && var2 && var3 && var4 && var5
}
func compareResources(spec1 v1alpha1.Resources, spec2 v1alpha1.Resources) bool {
	var1 := spec1.LimitCpu.Equal(spec2.LimitCpu)
	var2 := spec1.LimitMemory.Equal(spec2.LimitMemory)
	var3 := spec1.RequestCpu.Equal(spec2.RequestCpu)
	var4 := spec1.RequestMemory.Equal(spec2.RequestMemory)
	return var1 && var2 && var3 && var4
}
