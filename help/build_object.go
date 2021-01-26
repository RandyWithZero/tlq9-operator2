package help

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"tlq9-operator/api/v1alpha1"
)

func BuildDataVolume(spec *v1alpha1.Spec) (*v1.Volume, *v1.PersistentVolumeClaim) {
	if spec == nil || spec.DataPersistentSpec == nil {
		return nil, nil
	}
	switch spec.DataPersistentSpec.DataMountType {
	case v1alpha1.HostPath:
		return &v1.Volume{
			Name: "data",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: spec.DataPersistentSpec.HostPath,
				},
			},
		}, nil
	case v1alpha1.VolumeChaimTemplate:
		modes := make([]v1.PersistentVolumeAccessMode, 1)
		modes[0] = spec.DataPersistentSpec.AccessMode
		return nil, &v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: "data",
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: modes,
				Resources: v1.ResourceRequirements{
					Requests: map[v1.ResourceName]resource.Quantity{
						v1.ResourceStorage: spec.DataPersistentSpec.DataStorage,
					},
				},
			},
		}
	default:
		return nil, nil
	}
}
func BuildConfigVolume(spec *v1alpha1.Spec) []v1.Volume {
	if spec == nil {
		return nil
	}
	volumes := make([]v1.Volume, 0)
	if &spec.TopicConfigMapName != nil && spec.TopicConfigMapName != "" {
		volumes = append(volumes, v1.Volume{
			Name: "topic",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: spec.TopicConfigMapName,
					},
					Items: []v1.KeyToPath{
						{
							Key:  TopicConfigName,
							Path: TopicConfigName,
						},
					},
				},
			},
		})
	}
	if &spec.ZoneConfigMapName != nil && spec.ZoneConfigMapName != "" {
		volumes = append(volumes, v1.Volume{
			Name: "zone",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: spec.TopicConfigMapName,
					},
					Items: []v1.KeyToPath{
						{
							Key:  ZoneConfigName,
							Path: ZoneConfigName,
						},
					},
				},
			},
		})
	}
	return volumes
}
func BuildVolumeMounts(spec *v1alpha1.Spec) []v1.VolumeMount {
	if spec == nil {
		return nil
	}
	mounts := make([]v1.VolumeMount, 0)
	if spec.DataPersistentSpec != nil {
		data := v1.VolumeMount{
			Name:      "data",
			MountPath: spec.DataPersistentSpec.DataDir,
		}
		mounts = append(mounts, data)
	}
	if &spec.ZoneConfigMapName != nil && spec.ZoneConfigMapName != "" {
		mounts = append(mounts, v1.VolumeMount{
			Name:      "zone",
			ReadOnly:  true,
			MountPath: ZoneConfigDir,
			SubPath:   ZoneConfigName,
		})
	}
	if &spec.TopicConfigMapName != nil && spec.TopicConfigMapName != "" {
		mounts = append(mounts, v1.VolumeMount{
			Name:      "topic",
			ReadOnly:  true,
			MountPath: TopicConfigDir,
			SubPath:   TopicConfigName,
		})
	}
	return mounts
}
func BuildResourceRequirements(spec *v1alpha1.Spec) v1.ResourceRequirements {
	if spec == nil || spec.Resources == nil {
		return v1.ResourceRequirements{}
	}
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    spec.Resources.LimitCpu,
			v1.ResourceMemory: spec.Resources.LimitMemory,
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    spec.Resources.RequestCpu,
			v1.ResourceMemory: spec.Resources.RequestMemory,
		},
	}
}
