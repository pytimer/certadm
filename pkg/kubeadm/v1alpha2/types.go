package v1alpha2

type Configuration struct {
	CertificatesDir   string   `yaml:"certificatesDir"`
	APIServerCertSANs []string `yaml:"apiServerCertSANs,omitempty"`
	// NodeRegistration holds fields that relate to registering the new master node to the cluster
	NodeRegistration NodeRegistrationOptions `json:"nodeRegistration,omitempty"`
}

type NodeRegistrationOptions struct {
	Name string `yaml:"name,omitempty"`

	// CRISocket is used to retrieve container runtime info. This information will be annotated to the Node API object, for later re-use
	CRISocket string `yaml:"criSocket,omitempty"`
}