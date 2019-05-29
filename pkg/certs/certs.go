package certs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pytimer/certadm/pkg/constants"
	"github.com/pytimer/certadm/pkg/kubeadm"

	"github.com/otiai10/copy"
	"k8s.io/klog"
	"k8s.io/utils/path"
	"k8s.io/utils/temp"
)

var defaultCertificates = []string{
	"apiserver.crt",
	"apiserver.key",
	"apiserver-kubelet-client.crt",
	"apiserver-kubelet-client.key",
	// Front Proxy certs
	"front-proxy-ca.crt",
	"front-proxy-ca.key",
	"front-proxy-client.crt",
	"front-proxy-client.key",
	// etcd certs
	"etcd/healthcheck-client.crt",
	"etcd/healthcheck-client.key",
	"etcd/peer.crt",
	"etcd/peer.key",
	"etcd/server.crt",
	"etcd/server.key",
	"apiserver-etcd-client.crt",
	"apiserver-etcd-client.key",
}

// BackupCertificates ...
func BackupCertificates(src, dest string) error {
	if dest == "" {
		dir, err := temp.CreateTempDir(constants.TempDirPrefix)
		if err != nil {
			dest = constants.DefaultTempCertificatesDir
		} else {
			dest = dir.Name
		}
	}

	klog.V(2).Infof("[certs] Backup certificates from %s to %s \n", src, dest)
	if err := copy.Copy(src, dest); err != nil {
		return err
	}

	return nil
}

// RemoveOldCertificates remove unused certificates in certDir.
func RemoveOldCertificates(certDir string) error {
	for _, cert := range defaultCertificates {
		p := filepath.Join(certDir, cert)
		if exists, err := path.Exists(path.CheckFollowSymlink, p); err != nil {
			return err
		} else if !exists {
			continue
		}
		if err := os.Remove(p); err != nil {
			return err
		}
	}
	return nil
}

func RenewCertificate(conf string) error {
	if exists, err := path.Exists(path.CheckFollowSymlink, conf); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("config file: %s not exists", conf)
	}

	out, err := kubeadm.PhasesCreateCerts(conf)
	klog.Info(string(out))
	if err != nil {
		return err
	}

	return nil
}

func RemoveKubeletCertificate(certDir string) error {
	return filepath.Walk(certDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		return os.Remove(path)
	})
}
