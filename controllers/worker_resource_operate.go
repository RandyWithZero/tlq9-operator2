package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"
	"tlq9-operator/api/v1alpha1"
	"tlq9-operator/help"
)

type WorkerOperate struct {
	log logr.Logger
	r   *TLQWorkerReconciler
	ctx context.Context
	req ctrl.Request
}

func (o *WorkerOperate) GetWorker() (*v1alpha1.TLQWorker, ctrl.Result, error) {
	worker := &v1alpha1.TLQWorker{}
	err := o.r.Get(o.ctx, o.req.NamespacedName, worker)
	if err != nil {
		if errors.IsNotFound(err) {
			o.log.Info("TLQWorker resource not found. Ignoring since object must be deleted.")
			return nil, ctrl.Result{}, nil
		}
		o.log.Error(err, "Failed to get TLQWorker")
		return nil, ctrl.Result{}, err
	}
	return worker, ctrl.Result{}, nil
}
func (o *WorkerOperate) CreateOrUpdateService(worker *v1alpha1.TLQWorker) (*v1.Service, ctrl.Result, error) {
	service := &v1.Service{}
	err := o.r.Get(o.ctx, types.NamespacedName{Name: worker.Name, Namespace: worker.Namespace}, service)
	if err != nil {
		if errors.IsNotFound(err) {
			service := buildServiceInstanceForWorker(worker)
			if err := controllerutil.SetControllerReference(worker, service, o.r.Scheme); err != nil {
				return nil, ctrl.Result{}, err
			}
			o.log.Info("set service owner:" + worker.Name)
			if err := o.r.Create(o.ctx, service); err != nil && !errors.IsAlreadyExists(err) {
				return nil, ctrl.Result{}, err
			}
			o.log.Info("create reference service" + worker.Name)
		} else {
			return nil, ctrl.Result{}, err
		}
	} else {
		if service.Spec.Ports[0].Port != worker.Spec.Detail.Port {
			service.Spec.Ports[0].Port = worker.Spec.Detail.Port
			err := o.r.Update(o.ctx, service)
			o.log.Info("update reference service" + worker.Name)
			if err != nil {
				return nil, ctrl.Result{}, err
			} else {
				return service, ctrl.Result{}, nil
			}

		}
	}
	return service, ctrl.Result{}, nil
}
func (o *WorkerOperate) UpdateWorkerStatus(worker *v1alpha1.TLQWorker, statefulSet *v12.StatefulSet, service *v1.Service) (ctrl.Result, error) {
	o.log.Info("update TLQWorker resource status ...")
	oldStatus := worker.Status.Parse
	if statefulSet.Status.ReadyReplicas == defaultReplicas {
		worker.Status.Parse = v1alpha1.Healthy
		list := &v1.PodList{}
		labelSelector := labels.SelectorFromSet(statefulSet.Spec.Selector.MatchLabels)
		listOps := &client.ListOptions{
			Namespace:     statefulSet.Namespace,
			LabelSelector: labelSelector,
		}
		err := o.r.List(o.ctx, list, listOps)
		if err != nil {
			return ctrl.Result{}, err
		}
		pod := list.Items[0]
		worker.Status.Server = pod.Status.HostIP + ":" + strconv.Itoa(int(service.Spec.Ports[0].NodePort))
	} else if statefulSet.Status.ReadyReplicas == 0 {
		worker.Status.Parse = v1alpha1.Pending
		worker.Status.Server = ""
	}
	if oldStatus != worker.Status.Parse {
		if err := o.r.Status().Update(o.ctx, worker); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (o *WorkerOperate) CreateOrUpdateStatefulSet(worker *v1alpha1.TLQWorker, service *v1.Service) (*v12.StatefulSet, ctrl.Result, error) {
	statefulSet := &v12.StatefulSet{}
	err := o.r.Get(o.ctx, types.NamespacedName{Name: worker.Name, Namespace: worker.Namespace}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			forWorker := buildStatefulSetInstanceForWorker(worker)
			SetEnvForWorker(forWorker, service, worker)
			if err := controllerutil.SetControllerReference(forWorker, statefulSet, o.r.Scheme); err != nil {
				return nil, ctrl.Result{}, err
			}
			o.log.Info("set statefulSet owner ...")
			if err := o.r.Create(o.ctx, statefulSet); err != nil && !errors.IsAlreadyExists(err) {
				return nil, ctrl.Result{}, err
			}
			o.log.Info("create reference statefulSet...")
			return statefulSet, ctrl.Result{}, err
		} else {
			return nil, ctrl.Result{}, err
		}
	} else {
		statefulSetNew := buildStatefulSetInstanceForWorker(worker)
		annotations := statefulSetNew.Annotations
		SetEnvForWorker(statefulSetNew, service, worker)
		if !bytes.Equal([]byte(statefulSet.Annotations["owner-spec"]), []byte(statefulSetNew.Annotations["owner-spec"])) {
			o.log.Info("update reference statefulSet...")
			statefulSetNew.ObjectMeta = *statefulSet.ObjectMeta.DeepCopy()
			statefulSetNew.Annotations = annotations
			err := o.r.Update(o.ctx, statefulSetNew)
			if err != nil {
				return nil, ctrl.Result{}, err
			} else {
				return statefulSet, ctrl.Result{}, nil
			}

		} else {
			return statefulSet, ctrl.Result{}, nil
		}
	}

}
func buildServiceInstanceForWorker(worker *v1alpha1.TLQWorker) *v1.Service {
	ports := make([]v1.ServicePort, 1)
	ports[0] = v1.ServicePort{
		Name:       "worker-port",
		TargetPort: intstr.FromInt(int(worker.Spec.Detail.Port)),
		Port:       worker.Spec.Detail.Port,
	}
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      worker.Name,
			Namespace: worker.Namespace,
			Labels:    worker.Labels,
		},
		Spec: v1.ServiceSpec{
			Selector: worker.Labels,
			Type:     v1.ServiceTypeNodePort,
			Ports:    ports,
		},
	}
}
func buildStatefulSetInstanceForWorker(worker *v1alpha1.TLQWorker) *v12.StatefulSet {
	containers := make([]v1.Container, 1)
	ports := make([]v1.ContainerPort, 1)
	ports[0] = v1.ContainerPort{
		ContainerPort: worker.Spec.Detail.Port,
	}
	policy := worker.Spec.Detail.ImagePullPolicy
	if &policy == nil || "" == policy {
		policy = v1.PullAlways
	}
	volumes := help.BuildConfigVolume(worker.Spec.Detail)
	dataVolume, claimTemplate := help.BuildDataVolume(worker.Spec.Detail)
	var claimTemplates []v1.PersistentVolumeClaim
	if dataVolume != nil {
		volumes = append(volumes, *dataVolume)
	}
	if claimTemplate != nil {
		claimTemplates = []v1.PersistentVolumeClaim{}
		claimTemplates[0] = *claimTemplate
	}
	requirements := help.BuildResourceRequirements(worker.Spec.Detail)
	mounts := help.BuildVolumeMounts(worker.Spec.Detail)
	containers[0] = v1.Container{
		Name:            worker.Name,
		Image:           worker.Spec.Detail.Image,
		ImagePullPolicy: policy,
		VolumeMounts:    mounts,
		Ports:           ports,
		Resources:       requirements,
	}
	template := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      worker.Name,
			Namespace: worker.Namespace,
			Labels:    worker.Labels,
		},
		Spec: v1.PodSpec{
			Volumes:       volumes,
			Containers:    containers,
			RestartPolicy: v1.RestartPolicyAlways,
		},
	}
	masterJson, _ := json.Marshal(worker.Spec)
	statefulSet := &v12.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      worker.Name,
			Namespace: worker.Namespace,
			Labels:    worker.Labels,
			Annotations: map[string]string{
				"owner-spec": string(masterJson),
			},
		},
		Spec: v12.StatefulSetSpec{
			Replicas: &defaultReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: worker.Labels,
			},
			Template:             template,
			VolumeClaimTemplates: claimTemplates,
		},
	}
	return statefulSet
}

func SetEnvForWorker(statefulSet *v12.StatefulSet, service *v1.Service, worker *v1alpha1.TLQWorker) {
	clusterName := worker.Labels["cluster"]
	index := worker.Labels["index"]
	nameserver := worker.Annotations["nameserver"]
	envs := statefulSet.Spec.Template.Spec.Containers[0].Env
	nodePort := service.Spec.Ports[0].NodePort
	e1 := v1.EnvVar{
		Name:  "cluster_name",
		Value: clusterName,
	}
	e2 := v1.EnvVar{
		Name: "broker_name",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		},
	}
	e3 := v1.EnvVar{
		Name:  "brokerid",
		Value: index,
	}
	e4 := v1.EnvVar{
		Name:  "listen_port",
		Value: strconv.Itoa(int(nodePort)),
	}
	e5 := v1.EnvVar{
		Name: "listen_ip",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "status.hostIP",
			},
		},
	}
	e6 := v1.EnvVar{
		Name:  "register_status",
		Value: strconv.Itoa(worker.Spec.RegisterStatus),
	}
	e7 := v1.EnvVar{
		Name:  "register_listen_port",
		Value: strconv.Itoa(int(nodePort)),
	}
	e8 := v1.EnvVar{
		Name: "register_listen_ipv4",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "status.hostIP",
			},
		},
	}
	e9 := v1.EnvVar{
		Name:  "homedir",
		Value: worker.Spec.WorkRootDir,
	}
	e10 := v1.EnvVar{
		Name:  "log_level",
		Value: strconv.Itoa(worker.Spec.LogLevel),
	}
	e11 := v1.EnvVar{
		Name:  "is_affinity",
		Value: strconv.Itoa(worker.Spec.IsAffinity),
	}

	e12 := v1.EnvVar{
		Name:  "request_service_num",
		Value: strconv.Itoa(worker.Spec.RequestServiceNum),
	}
	e13 := v1.EnvVar{
		Name:  "response_service_num",
		Value: strconv.Itoa(worker.Spec.ResponseServiceNum),
	}
	e14 := v1.EnvVar{
		Name:  "nameserver_url_list",
		Value: nameserver,
	}
	if envs == nil || cap(envs) == 0 {
		envs = make([]v1.EnvVar, 14)
		envs[0] = e1
		envs[1] = e2
		envs[2] = e3
		envs[3] = e4
		envs[4] = e5
		envs[5] = e6
		envs[6] = e7
		envs[7] = e8
		envs[8] = e9
		envs[9] = e10
		envs[10] = e11
		envs[11] = e12
		envs[12] = e13
		envs[13] = e14

	} else {
		envs = append(envs, e1, e2, e3, e4, e5, e6, e7, e8, e9, e10, e11, e12, e13, e14)
	}
	statefulSet.Spec.Template.Spec.Containers[0].Env = envs
}
