package data

import (
	"context"
	hdfsv1alpha1 "github.com/zncdata-labs/hdfs-operator/api/v1alpha1"
	"github.com/zncdata-labs/hdfs-operator/internal/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceReconciler struct {
	common.GeneralResourceStyleReconciler[*hdfsv1alpha1.HdfsCluster, *hdfsv1alpha1.RoleGroupSpec]
}

// NewService new a ServiceReconciler
func NewServiceHeadless(
	scheme *runtime.Scheme,
	instance *hdfsv1alpha1.HdfsCluster,
	client client.Client,
	groupName string,
	labels map[string]string,
	mergedCfg *hdfsv1alpha1.RoleGroupSpec,
) *ServiceReconciler {
	return &ServiceReconciler{
		GeneralResourceStyleReconciler: *common.NewGeneraResourceStyleReconciler(
			scheme,
			instance,
			client,
			groupName,
			labels,
			mergedCfg,
		),
	}
}

func (s *ServiceReconciler) Build(_ context.Context) (client.Object, error) {
	serviceType := common.HeadlessService
	return common.NewServiceBuilder(
		createServiceName(s.Instance.GetName(), s.GroupName),
		s.Instance.GetNamespace(),
		s.MergedLabels,
		s.makePorts(),
	).SetClusterIP(&serviceType).Build(), nil
}

func (s *ServiceReconciler) makePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:       hdfsv1alpha1.MetricName,
			Port:       ServiceMetricPort,
			Protocol:   corev1.ProtocolTCP,
			TargetPort: intstr.FromString(hdfsv1alpha1.MetricName),
		},
		{
			Name:       hdfsv1alpha1.DataName,
			Port:       ServiceDataPort,
			Protocol:   corev1.ProtocolTCP,
			TargetPort: intstr.FromString(hdfsv1alpha1.DataName),
		},
		{
			Name:       hdfsv1alpha1.IpcName,
			Port:       ServiceIpcPort,
			Protocol:   corev1.ProtocolTCP,
			TargetPort: intstr.FromString(hdfsv1alpha1.IpcName),
		},
		{
			Name:       hdfsv1alpha1.HttpName,
			Port:       ServiceHttpPort,
			Protocol:   corev1.ProtocolTCP,
			TargetPort: intstr.FromString(hdfsv1alpha1.HttpName),
		},
	}
}
