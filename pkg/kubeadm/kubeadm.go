package kubeadm

import (
	"fmt"
	stdexec "os/exec"

	"github.com/pytimer/certadm/pkg/constants"
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

func getKubeadmAPIVersion(version string) string {
	return "v1alpha2"
}

func PhasesAlphaCerts(configFile string) ([]byte, error) {
	kubeadmVersion := constants.DefaultKubeadmVersion
	c := exec.New().Command(kubeadmExecPath, "version", "-o", "short")
	b, err := c.Output()
	if err == nil {
		kubeadmVersion = string(b)
	}


	args := []string{
		"alpha",
		"phase",
		"certs",
		"all",
		fmt.Sprintf("--config=%s", configFile),
	}
	cmd := exec.New().Command(kubeadmExecPath, args...)
	return cmd.CombinedOutput()
}
