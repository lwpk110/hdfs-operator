package journal

import (
	"context"
	hdfsv1alpha1 "github.com/zncdata-labs/hdfs-operator/api/v1alpha1"
	"github.com/zncdata-labs/hdfs-operator/internal/common"
	"github.com/zncdata-labs/hdfs-operator/internal/controller/data/container"
	"github.com/zncdata-labs/hdfs-operator/internal/util"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatefulSetReconciler struct {
	common.WorkloadStyleReconciler[*hdfsv1alpha1.HdfsCluster, *hdfsv1alpha1.RoleGroupSpec]
}

func NewStatefulSet(
	scheme *runtime.Scheme,
	instance *hdfsv1alpha1.HdfsCluster,
	client client.Client,
	groupName string,
	labels map[string]string,
	mergedCfg *hdfsv1alpha1.RoleGroupSpec,
	replicate int32,
) *StatefulSetReconciler {
	return &StatefulSetReconciler{
		WorkloadStyleReconciler: *common.NewWorkloadStyleReconciler(
			scheme,
			instance,
			client,
			groupName,
			labels,
			mergedCfg,
			replicate,
		),
	}
}

func (s *StatefulSetReconciler) Build(_ context.Context) (client.Object, error) {
	return &appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      createStatefulSetName(s.Instance.GetName(), s.GroupName),
			Namespace: s.Instance.GetNamespace(),
			Labels:    s.MergedLabels,
		},
		Spec: appv1.StatefulSetSpec{
			Replicas: s.getReplicates(),
			Selector: &metav1.LabelSelector{MatchLabels: s.MergedLabels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: s.MergedLabels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: common.CreateServiceAccountName(s.Instance.GetName()),
					SecurityContext:    s.MergedCfg.Config.SecurityContext,
					Containers: []corev1.Container{
						s.makeJournalNodeContainer(),
					},
					Volumes: s.makeVolumes(),
				},
			},
			VolumeClaimTemplates: s.makePvcTemplates(),
		},
	}, nil
}
func (s *StatefulSetReconciler) CommandOverride(resource client.Object) {
	//TODO implement me
	//panic("implement me")
}

func (s *StatefulSetReconciler) EnvOverride(resource client.Object) {
	//TODO implement me
	//panic("implement me")
}

func (s *StatefulSetReconciler) LogOverride(resource client.Object) {
	//TODO implement me
	//panic("implement me")
}

// make name node container
func (s *StatefulSetReconciler) makeJournalNodeContainer() corev1.Container {
	image := s.getImageSpec()
	journalNode := NewJournalNodeContainerBuilder(
		util.ImageRepository(image.Repository, image.Tag),
		image.PullPolicy,
		*util.ConvertToResourceRequirements(s.MergedCfg.Config.Resources),
		s.getZookeeperDiscoveryZNode(),
	)
	return journalNode.Build(journalNode)
}

// make volumes
func (s *StatefulSetReconciler) makeVolumes() []corev1.Volume {
	limit := resource.MustParse("150Mi")
	return []corev1.Volume{
		{
			Name: logVolumeName(),
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					SizeLimit: &limit,
				},
			},
		},
		{
			Name: journalNodeConfigVolumeName(),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: s.getConfigMapSource(),
			},
		},
		{
			Name: journalNodeLogVolumeName(),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: s.getConfigMapSource(),
			},
		},
	}
}

func (s *StatefulSetReconciler) makePvcTemplates() []corev1.PersistentVolumeClaim {
	return []corev1.PersistentVolumeClaim{
		s.createDataPvcTemplate(),
	}
}

// create data pvc template
func (s *StatefulSetReconciler) createDataPvcTemplate() corev1.PersistentVolumeClaim {
	return corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: container.DataVolumeName(),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}
}

func (s *StatefulSetReconciler) getReplicates() *int32 {
	return &s.MergedCfg.Replicas
}

// get image spec
func (s *StatefulSetReconciler) getImageSpec() *hdfsv1alpha1.ImageSpec {
	return s.Instance.Spec.Image
}

func (s *StatefulSetReconciler) getConfigMapSource() *corev1.ConfigMapVolumeSource {
	return &corev1.ConfigMapVolumeSource{
		LocalObjectReference: corev1.LocalObjectReference{
			Name: createConfigName(s.Instance.GetName(), s.GroupName)}}
}

// get zookeeper discovery znode
func (s *StatefulSetReconciler) getZookeeperDiscoveryZNode() string {
	return s.Instance.Spec.ClusterConfigSpec.ZookeeperDiscoveryZNode
}
