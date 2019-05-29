package v1alpha2

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/pytimer/certadm/pkg/kubeadm"
)

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

func (k *KubeadmAlpha2) LoadConfigFromFile(f string) (*kubeadm.Config, error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	c := &Configuration{}
	if err := yaml.Unmarshal(b, c); err != nil {
		return nil, err
	}

	return &kubeadm.Config{
		CertificatesDir: c.CertificatesDir,
		CRISocket: c.NodeRegistration.CRISocket,
	}, nil
}