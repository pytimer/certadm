package kubeadm

type Config struct {
	CertificatesDir   string   `yaml:"certificatesDir"`
	APIServerCertSANs []string `yaml:"apiServerCertSANs,omitempty"`
	CRISocket         string   `yaml:"criSocket,omitempty"`
}
