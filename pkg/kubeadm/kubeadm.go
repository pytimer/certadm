package kubeadm

import (
	"fmt"
	stdexec "os/exec"
	"strings"

	"github.com/pytimer/certadm/pkg/constants"

	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/klog"
	"k8s.io/utils/exec"
)

var kubeadmExecPath string

func init() {
	var err error
	kubeadmExecPath, err = stdexec.LookPath("kubeadm")
	if err != nil {
		kubeadmExecPath = "/usr/bin/kubeadm"
	}
}

func compareSemanticVersion(v1 *version.Version, v2 string) int {
	res, err := v1.Compare(v2)
	if err != nil {
		return 0
	}
	return res
}

func getKubeadmAPIVersion(v string) string {
	semanticVersion := version.MustParseGeneric(v)
	switch {
	case compareSemanticVersion(semanticVersion, "v1.11.0") >= 0 && compareSemanticVersion(semanticVersion, "v1.12.0") < 0:
		return "v1alpha2"
	case compareSemanticVersion(semanticVersion, "v1.12.0") >= 0 && compareSemanticVersion(semanticVersion, "v1.13.0") < 0:
		return "v1alpha3"
	case compareSemanticVersion(semanticVersion, "v1.13.0") >= 0 && compareSemanticVersion(semanticVersion, "v1.15.0") < 0:
		return "v1beta1"
	case compareSemanticVersion(semanticVersion, "v1.15.0") >= 0:
		return "v1beta2"
	}

	return constants.DefaultKubeadmAPIVersion
}

func getKubeadmVersion() string {
	klog.V(3).Info("get kubeadm version")
	kubeadmVersion := constants.DefaultKubeadmVersion
	c := exec.New().Command(kubeadmExecPath, "version", "-o", "short")
	b, err := c.Output()
	if err == nil {
		kubeadmVersion = strings.TrimRight(string(b), "\n")
	}
	return kubeadmVersion
}

// GetKubeadmAPIVersion returns the kubeadm version via `kubeadm version -o short`
func GetKubeadmAPIVersion() string {
	v := getKubeadmVersion()
	klog.V(1).Infof("[kubeadm] Using kubeadm version is %s\n", v)
	return getKubeadmAPIVersion(v)
}

func PhasesCreateCerts(configFile string) ([]byte, error) {
	kubeadmAPIVersion := GetKubeadmAPIVersion()
	args := GetKubeadmFactory(kubeadmAPIVersion).RenewCertsCommandArgs()
	args = append(args, fmt.Sprintf("--config=%s", configFile))
	klog.V(2).Infof("[kubeadm-certs] renew certificates command args: '%s %s'", kubeadmExecPath, strings.Join(args, " "))
	cmd := exec.New().Command(kubeadmExecPath, args...)
	return cmd.CombinedOutput()
}

func PhasesCreateKubeConfig(configFile string) ([]byte, error) {
	kubeadmAPIVersion := GetKubeadmAPIVersion()
	args := GetKubeadmFactory(kubeadmAPIVersion).RenewKubeConfigCommandArgs()
	args = append(args, fmt.Sprintf("--config=%s", configFile))
	klog.V(2).Infof("[kubeadm-kubeconfig] renew kubeconfig command args: '%s %s'", kubeadmExecPath, strings.Join(args, " "))
	cmd := exec.New().Command(kubeadmExecPath, args...)
	return cmd.CombinedOutput()
}

// FetchConfigurationFromConfigFile returns the configurations from the kubeadm config file
func FetchConfigurationFromConfigFile(configFile string) (*Config, error) {
	v := GetKubeadmAPIVersion()
	factory := GetKubeadmFactory(v)

	c, err := factory.LoadConfigFromFile(configFile)
	return c, err
}

// GetCertificatesDirFromConfigFile returns the certificates directory from the kubeadm config file
func GetCertificatesDirFromConfigFile(configFile string) (string, error) {
	c, err := FetchConfigurationFromConfigFile(configFile)
	if err != nil {
		return "", err
	}

	return c.CertificatesDir, nil
}