package v1alpha2

import "github.com/pytimer/certadm/pkg/kubeadm"

func init() {
	kubeadm.Add("v1.11")
}

func RenewCertsCommandArgs() []string {
	args := []string{
		"alpha",
		"phase",
		"certs",
		"all",
	}
	return args
}