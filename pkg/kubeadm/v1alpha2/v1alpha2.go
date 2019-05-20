package v1alpha2

import "github.com/pytimer/certadm/pkg/kubeadm"

func init() {
	kubeadm.Add("v1alpha2", NewKubeadmAlpha2())
}

type KubeadmAlpha2 struct {
}

func NewKubeadmAlpha2() *KubeadmAlpha2 {
	return &KubeadmAlpha2{}
}

func (k *KubeadmAlpha2) RenewCertsCommandArgs() []string {
	args := []string{
		"alpha",
		"phase",
		"certs",
		"all",
	}
	return args
}

func (k *KubeadmAlpha2) RenewKubeConfigCommandArgs() []string {
	args := []string{
		"alpha",
		"phase",
		"kubeconfig",
		"all",
	}
	return args
}
