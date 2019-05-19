package kubeadm

var versions = make(map[string]string)

func Add(key, version string) {
	versions[key] = version
}

func GetAPIVersion(key string) string {
	v, ok := versions[key]
	if !ok {
		return "v1alpha2"
	}
	return v
}
