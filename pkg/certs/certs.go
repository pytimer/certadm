package certs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pytimer/certadm/pkg/constants"

	"github.com/otiai10/copy"
	"k8s.io/klog"
	ec "k8s.io/utils/exec"
	"k8s.io/utils/path"
	"k8s.io/utils/temp"
)

var excludeCertificates = []string{
	"ca.crt",
	"ca.key",
	"sa.key",
	"sa.pub",
	"etcd/ca.crt",
	"etcd/ca.key",
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

	if err := copy.Copy(src, dest); err != nil {
		return err
	}

	return nil
}

// RemoveOldCertificates remove unused certificates in certDir.
func RemoveOldCertificates(certDir string) error {
	return filepath.Walk(certDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if checkCertificateIsDeleted(strings.Split(path, certDir+"/")[1]) {
			if err := os.Remove(path); err != nil {
				return err
			}
		}

		return nil
	})
}

func checkCertificateIsDeleted(cert string) bool {
	for _, v := range excludeCertificates {
		if strings.EqualFold(cert, v) {
			return false
		}
	}
	return true
}

func RenewCertificate(conf string) error {
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
		"certs",
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

func RemoveKubeletCertificate(certDir string) error {
	return filepath.Walk(certDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		return os.Remove(path)
	})
}