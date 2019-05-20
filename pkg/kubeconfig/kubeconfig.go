package kubeconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/klog"
	"k8s.io/utils/path"

	"github.com/pytimer/certadm/pkg/kubeadm"
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

	out, err := kubeadm.PhasesCreateKubeConfig(conf)
	klog.Info(string(out))
	if err != nil {
		return err
	}

	return nil
}
