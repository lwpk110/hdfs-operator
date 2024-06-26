package container

import (
	hdfsv1alpha1 "github.com/zncdata-labs/hdfs-operator/api/v1alpha1"
	"github.com/zncdata-labs/hdfs-operator/internal/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type NameNodeContainerBuilder struct {
	common.ContainerBuilder
	zookeeperDiscoveryZNode string
}

func NewNameNodeContainerBuilder(
	image string,
	imagePullPolicy corev1.PullPolicy,
	resource corev1.ResourceRequirements,
	zookeeperDiscoveryZNode string,
) *NameNodeContainerBuilder {
	return &NameNodeContainerBuilder{
		ContainerBuilder:        *common.NewContainerBuilder(image, imagePullPolicy, resource),
		zookeeperDiscoveryZNode: zookeeperDiscoveryZNode,
	}
}

// container name
func (n *NameNodeContainerBuilder) ContainerName() string {
	return string(NameNode)
}

func (n *NameNodeContainerBuilder) Command() []string {
	return common.GetCommonCommand()
}

func (n *NameNodeContainerBuilder) ContainerEnv() []corev1.EnvVar {
	envs := common.GetCommonContainerEnv(n.zookeeperDiscoveryZNode, NameNode)
	envs = append(envs, corev1.EnvVar{
		Name:  "HDFS_NAMENODE_OPTS",
		Value: "-Djava.security.properties=/znclabs/config/namenode/security.properties -Xmx838860k",
	})
	return envs
}

func (n *NameNodeContainerBuilder) VolumeMount() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      LogVolumeName(),
			MountPath: "/znclabs/log",
		},
		{
			Name:      NameNodeConfVolumeName(),
			MountPath: "/znclabs/mount/config/namenode",
		},
		{
			Name:      NameNodeLogVolumeName(),
			MountPath: "/znclabs/mount/log/namenode",
		},
		{
			Name:      ListenerVolumeName(),
			MountPath: "/znclabs/listener",
		},
		{
			Name:      DataVolumeName(),
			MountPath: "/znclabs/data",
		},
	}
}

func (n *NameNodeContainerBuilder) LivenessProbe() *corev1.Probe {
	return &corev1.Probe{
		FailureThreshold:    5,
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		TimeoutSeconds:      1,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/dfshealth.html",
				Port:   intstr.FromString(hdfsv1alpha1.HttpName),
				Scheme: corev1.URISchemeHTTP,
			},
		},
	}
}

func (n *NameNodeContainerBuilder) ReadinessProbe() *corev1.Probe {
	return &corev1.Probe{
		FailureThreshold:    3,
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		TimeoutSeconds:      1,
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromString(hdfsv1alpha1.RpcName)},
		},
	}

}

// ContainerPorts  make container ports of name node
func (n *NameNodeContainerBuilder) ContainerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          hdfsv1alpha1.HttpName,
			ContainerPort: hdfsv1alpha1.NameNodeHttpPort,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          hdfsv1alpha1.RpcName,
			ContainerPort: hdfsv1alpha1.NameNodeRpcPort,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          hdfsv1alpha1.MetricName,
			ContainerPort: hdfsv1alpha1.NameNodeMetricPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}
}

func (n *NameNodeContainerBuilder) CommandArgs() []string {
	return []string{
		`mkdir -p /znclabs/config/namenode
cp /znclabs/mount/config/namenode/*.xml /znclabs/config/namenode
cp /znclabs/mount/config/namenode/namenode.log4j.properties /znclabs/config/namenode/log4j.properties
\

prepare_signal_handlers()
{
  unset term_child_pid
  unset term_kill_needed
  trap 'handle_term_signal' TERM
}

handle_term_signal()
{
  if [ "${term_child_pid}" ]; then
	  kill -TERM "${term_child_pid}" 2>/dev/null
  else
	  term_kill_needed="yes"
  fi
}

wait_for_termination()
{
  set +e
  term_child_pid=$1
  if [[ -v term_kill_needed ]]; then
	  kill -TERM "${term_child_pid}" 2>/dev/null
  fi
  wait ${term_child_pid} 2>/dev/null
  trap - TERM
  wait ${term_child_pid} 2>/dev/null
  set -e
}

rm -f /znclabs/log/_vector/shutdown
prepare_signal_handlers
if [[ -d /znclabs/listener ]]; then
  export POD_ADDRESS=$(cat /znclabs/listener/default-address/address)
  for i in /znclabs/listener/default-address/ports/*; do
	  export $(basename $i | tr a-z A-Z)_PORT="$(cat $i)"
  done
fi
/znclabs/hadoop/bin/hdfs namenode &
wait_for_termination $!
mkdir -p /znclabs/log/_vector && touch /znclabs/log/_vector/shutdown
`,
	}
}
