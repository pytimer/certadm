package main

import (
	"path/filepath"

	"github.com/pytimer/certadm/pkg/certs"
	"github.com/pytimer/certadm/pkg/constants"
	"github.com/pytimer/certadm/pkg/kubeconfig"

	"github.com/spf13/cobra"
	"k8s.io/klog"
	"github.com/otiai10/copy"
)

var (
	KubernetesDir    = "/etc/kubernetes"
	kubernetesPkiDir = "/etc/kubernetes/pki"
)

type renewOptions struct {
	KubeDir         string
	CertificatesDir string
	configFile      string
}

// NewCmdRenew returns "certadm renew" command.
func NewCmdRenew() *cobra.Command {
	opts := &renewOptions{
		KubeDir:         KubernetesDir,
		CertificatesDir: kubernetesPkiDir,
	}
	cmd := &cobra.Command{
		Use:   "renew",
		Short: "Run this command in order to renew Kubernetes cluster certificates",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.run(); err != nil {
				klog.Error(err)
				return
			}
		},
	}

	cmd.Flags().StringVar(&opts.KubeDir, "root-dir", KubernetesDir, "The Kubernetes root directory.")
	cmd.Flags().StringVar(&opts.CertificatesDir, "cert-dir", kubernetesPkiDir, "The path where to save and store the certificates.")
	cmd.Flags().StringVar(&opts.configFile, "config", "", "Using the config file to renew certificates.")

	return cmd
}

func (o *renewOptions) run() error {
	// 1. backup old certificates to temp dir.
	if err := certs.BackupCertificates(o.CertificatesDir, ""); err != nil {
		return err
	}

	// 2. remove old certificates exclude CA and sa.
	if err := certs.RemoveOldCertificates(o.CertificatesDir); err != nil {
		return err
	}

	// 3. remove old kubeconfig
	if err := kubeconfig.RemoveOldKubeconfig(o.KubeDir); err != nil {
		return err
	}

	// 4. renew certificates and kubeconfig
	if err := certs.RenewCertificate(o.configFile); err != nil {
		return err
	}

	if err := kubeconfig.RenewKubeConfigFile(o.configFile); err != nil {
		return err
	}

	// 5. remove kubelet certificates
	if err := certs.RemoveKubeletCertificate(constants.KubeletCertificatesPath); err != nil {
		return err
	}

	// 6. copy new admin.conf to $HOME/.kube/config
	if err := copy.Copy(filepath.Join(o.KubeDir,"admin.conf"), "~/.kube/config"); err != nil {
		return err
	}

	return nil
}
