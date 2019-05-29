package constants

import "time"

const (
	TempDirPrefix              = "certadm"
	DefaultTempCertificatesDir = "/tmp/certadm"
	KubeletCertificatesPath    = "/var/lib/kubelet/pki"

	DefaultKubeadmVersion    = "v1.11.0"
	DefaultKubeadmAPIVersion = "v1alpha2"

	KubernetesDir    = "/etc/kubernetes"

	ContainerCallRetryInterval = 10 * time.Second
	ContainerCallTimeout       = 5 * time.Minute

	ServiceCallRetryInterval = 5 * time.Second
	ServiceCallTimeout = 1 * time.Minute

	// DefaultDockerCRISocket defines the default Docker CRI socket
	DefaultDockerCRISocket = "/var/run/dockershim.sock"
)

var ControlPlaneNames = []string{
	"kube-apiserver",
	"kube-scheduler",
	"kube-controller-manager",
	"etcd",
}