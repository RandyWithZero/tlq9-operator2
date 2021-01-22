package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strconv"
	"tlq9-operator/api/v1alpha1"
)

var (
	defaultReplicas int32 = 1
	defaultLabels         = map[string]string{"role": "master"}
)

type MasterOperate struct {
	log logr.Logger
	r   *TLQMasterReconciler
	ctx context.Context
	req ctrl.Request
}

func (o *MasterOperate) GetMaster() (*v1alpha1.TLQMaster, ctrl.Result, error) {
	master := &v1alpha1.TLQMaster{}
	err := o.r.Get(o.ctx, o.req.NamespacedName, master)
	if err != nil {
		if errors.IsNotFound(err) {
			o.log.Info("TLQMaster resource not found. Ignoring since object must be deleted.")
			return nil, ctrl.Result{}, nil
		}
		o.log.Error(err, "Failed to get TLQMaster")
		return nil, ctrl.Result{}, err
	}
	return master, ctrl.Result{}, nil
}
func (o *MasterOperate) CreateOrUpdateService(master *v1alpha1.TLQMaster) (*v1.Service, ctrl.Result, error) {
	service := &v1.Service{}
	err := o.r.Get(o.ctx, types.NamespacedName{Name: master.Name, Namespace: master.Namespace}, service)
	if err != nil {
		if errors.IsNotFound(err) {
			service := buildServiceInstance(master)
			if err := controllerutil.SetControllerReference(master, service, o.r.Scheme); err != nil {
				return nil, ctrl.Result{}, err
			}
			o.log.Info("set service owner ...")
			if err := o.r.Create(o.ctx, service); err != nil && !errors.IsAlreadyExists(err) {
				return nil, ctrl.Result{}, err
			}
			o.log.Info("create reference service...")
		} else {
			return nil, ctrl.Result{}, err
		}
	} else {
		if service.Spec.Ports[0].Port != master.Spec.Port {
			service.Spec.Ports[0].Port = master.Spec.Port
			err := o.r.Update(o.ctx, service)
			if err != nil {
				return nil, ctrl.Result{}, err
			} else {
				return service, ctrl.Result{}, nil
			}

		}
	}
	return service, ctrl.Result{}, nil
}
func (o *MasterOperate) UpdateMasterStatus(master *v1alpha1.TLQMaster, statefulSet *v12.StatefulSet) (ctrl.Result, error) {
	o.log.Info("update TLQMaster resource status ...")
	oldStatus := master.Status.Parse
	if statefulSet.Status.ReadyReplicas == defaultReplicas {
		master.Status.Parse = v1alpha1.Healthy
	} else if statefulSet.Status.ReadyReplicas == 0 {
		master.Status.Parse = v1alpha1.Pending
	}
	if oldStatus != master.Status.Parse {
		if err := o.r.Status().Update(o.ctx, master); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (o *MasterOperate) CreateOrUpdateStatefulSet(master *v1alpha1.TLQMaster, service *v1.Service) (*v12.StatefulSet, ctrl.Result, error) {
	statefulSet := &v12.StatefulSet{}
	err := o.r.Get(o.ctx, types.NamespacedName{Name: master.Name, Namespace: master.Namespace}, statefulSet)
	if err != nil {
		if errors.IsNotFound(err) {
			statefulSet := buildStatefulSetInstance(master)
			SetEnv(statefulSet, service, master)
			if err := controllerutil.SetControllerReference(master, statefulSet, o.r.Scheme); err != nil {
				return nil, ctrl.Result{}, err
			}
			o.log.Info("set statefulSet owner ...")
			if err := o.r.Create(o.ctx, statefulSet); err != nil && !errors.IsAlreadyExists(err) {
				return nil, ctrl.Result{}, err
			}
			o.log.Info("create reference statefulSet...")
		} else {
			return nil, ctrl.Result{}, err
		}
	} else {
		spec := statefulSet.Spec.Template.Spec
		statefulSetOld := buildStatefulSetInstance(&v1alpha1.TLQMaster{
			TypeMeta:   master.TypeMeta,
			ObjectMeta: *master.ObjectMeta.DeepCopy(),
			Spec: v1alpha1.TLQMasterSpec{
				Image:                spec.Containers[0].Image,
				ImagePullPolicy:      spec.Containers[0].ImagePullPolicy,
				Volumes:              spec.Volumes,
				VolumeClaimTemplates: statefulSet.Spec.VolumeClaimTemplates,
				Port:                 spec.Containers[0].Ports[0].ContainerPort,
				Envs:                 spec.Containers[0].Env,
				VolumeMounts:         spec.Containers[0].VolumeMounts,
			},
		})
		statefulSetNew := buildStatefulSetInstance(master)
		SetEnv(statefulSetNew, service, master)
		if !reflect.DeepEqual(&statefulSetNew, &statefulSetOld) {
			o.log.Info("update reference statefulSet...")
			statefulSetNew.ObjectMeta = *statefulSet.ObjectMeta.DeepCopy()
			marshal, _ := json.Marshal(statefulSetNew)
			fmt.Println(string(marshal))

			err := o.r.Update(o.ctx, statefulSetNew)
			if err != nil {
				return nil, ctrl.Result{}, err
			} else {
				return statefulSetOld, ctrl.Result{}, nil
			}

		}
	}
	return statefulSet, ctrl.Result{}, nil
}
func buildServiceInstance(master *v1alpha1.TLQMaster) *v1.Service {
	ports := make([]v1.ServicePort, 1)
	ports[0] = v1.ServicePort{
		Name:       "master-port",
		TargetPort: intstr.FromInt(int(master.Spec.Port)),
		Port:       master.Spec.Port,
	}
	defaultLabels["master"] = master.Name
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      master.Name,
			Namespace: master.Namespace,
			Labels:    defaultLabels,
		},
		Spec: v1.ServiceSpec{
			Selector: defaultLabels,
			Type:     v1.ServiceTypeNodePort,
			Ports:    ports,
		},
	}
}
func buildStatefulSetInstance(master *v1alpha1.TLQMaster) *v12.StatefulSet {
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
		Name:            master.Name,
		Image:           master.Spec.Image,
		ImagePullPolicy: policy,
		VolumeMounts:    master.Spec.VolumeMounts,
		Ports:           ports,
		Env:             master.Spec.Envs,
		Resources:       master.Spec.Resource,
	}
	defaultLabels["master"] = master.Name
	template := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:      master.Name,
			Namespace: master.Namespace,
			Labels:    defaultLabels,
		},
		Spec: v1.PodSpec{
			Volumes:       master.Spec.Volumes,
			Containers:    containers,
			RestartPolicy: v1.RestartPolicyAlways,
		},
	}
	statefulSet := &v12.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      master.Name,
			Namespace: master.Namespace,
			Labels:    defaultLabels,
		},
		Spec: v12.StatefulSetSpec{
			Replicas: &defaultReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: defaultLabels,
			},
			Template:             template,
			VolumeClaimTemplates: master.Spec.VolumeClaimTemplates,
		},
	}
	return statefulSet
}

func SetEnv(statefulSet *v12.StatefulSet, service *v1.Service, master *v1alpha1.TLQMaster) {
	marshal, _ := json.Marshal(service)
	fmt.Println(string(marshal))
	envs := statefulSet.Spec.Template.Spec.Containers[0].Env
	nodePort := service.Spec.Ports[0].NodePort
	e1 := v1.EnvVar{
		Name: "IpAddress",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "status.hostIP",
			},
		},
	}
	e2 := v1.EnvVar{
		Name:  "ListenPort",
		Value: strconv.Itoa(int(nodePort)),
	}
	e3 := v1.EnvVar{
		Name: "NameServerName",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		},
	}
	e4 := v1.EnvVar{
		Name:  "NameServerId",
		Value: "0",
	}
	e5 := v1.EnvVar{
		Name:  "UserName",
		Value: master.Spec.UserName,
	}
	e6 := v1.EnvVar{
		Name:  "Password",
		Value: master.Spec.Password,
	}
	e7 := v1.EnvVar{
		Name:  "VRRPPasswd",
		Value: master.Spec.VRRPPasswd,
	}
	e8 := v1.EnvVar{
		Name:  "AdvertiseInterval",
		Value: strconv.Itoa(int(master.Spec.AdvertiseInterval)),
	}
	if envs == nil || cap(envs) == 0 {
		envs = make([]v1.EnvVar, 8)
		envs[0] = e1
		envs[1] = e2
		envs[2] = e3
		envs[3] = e4
		envs[4] = e5
		envs[5] = e6
		envs[6] = e7
		envs[7] = e8

	} else {
		envs = append(envs, e1, e2, e3, e4, e5, e6, e7, e8)
	}
	statefulSet.Spec.Template.Spec.Containers[0].Env = envs
}
