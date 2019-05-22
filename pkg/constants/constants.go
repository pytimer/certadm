package constants

import "time"

const (
	TempDirPrefix              = "certadm"
	DefaultTempCertificatesDir = "/tmp/certadm"
	KubeletCertificatesPath    = "/var/lib/kubelet/pki"

	DefaultKubeadmVersion    = "v1.11.0"
	DefaultKubeadmAPIVersion = "v1alpha2"

	KubernetesDir    = "/etc/kubernetes"
	KubernetesPkiDir = "/etc/kubernetes/pki"

	ContainerCallRetryInterval = 10 * time.Second
	ContainerCallTimeout       = 5 * time.Minute
)
