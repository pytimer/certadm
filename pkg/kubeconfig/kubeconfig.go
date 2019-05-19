package kubeconfig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"k8s.io/klog"
	ec "k8s.io/utils/exec"
	"k8s.io/utils/path"
)

var kubeConfigs = []string{
	"admin.conf",
	"controller-manager.conf",
	"scheduler.conf",
	"kubelet.conf",
}

func RemoveOldKubeconfig(kubeconfigDir string) error {
	for _, kf := range kubeConfigs {
		kubeconfigPath := filepath.Join(kubeconfigDir, kf)
		if exists, err := path.Exists(path.CheckFollowSymlink, kubeconfigPath); err != nil {
			continue
		} else if !exists {
			continue
		}
		if err := os.Remove(kubeconfigPath); err != nil {
			return err
		}
	}
	return nil
}

func RenewKubeConfigFile(conf string) error {
	if exists, err := path.Exists(path.CheckFollowSymlink, conf); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("config file: %s not exists", conf)
	}

	kubeadmPath, err := exec.LookPath("kubeadm")
	if err != nil {
		return err
	}

	args := []string{
		"alpha",
		"phase",
		"kubeconfig",
		"all",
		fmt.Sprintf("--config=%s", conf),
	}
	cmd := ec.New().Command(kubeadmPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	klog.Info(string(out))

	return nil
}
