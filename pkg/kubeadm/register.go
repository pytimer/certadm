package kubeadm

var versions = make(map[string]Factory)

type Factory interface {
	RenewCertsCommandArgs() []string
	RenewKubeConfigCommandArgs() []string
}

func Add(key string, k Factory) {
	versions[key] = k
}

func GetKubeadmFactory(key string) Factory {
	v, ok := versions[key]
	if !ok {
		return nil
	}
	return v
}
