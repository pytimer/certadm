package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pytimer/certadm/pkg/certs"
	"github.com/pytimer/certadm/pkg/constants"
	"github.com/pytimer/certadm/pkg/kubeadm"
	"github.com/pytimer/certadm/pkg/kubeconfig"

	"github.com/spf13/cobra"
	"k8s.io/klog"
)

type renewOptions struct {
	kubernetesDir string
	configFile    string
}

// NewCmdRenew returns "certadm renew" command.
func NewCmdRenew() *cobra.Command {
	opts := &renewOptions{
		kubernetesDir: constants.KubernetesDir,
	}
	cmd := &cobra.Command{
		Use:   "renew",
		Short: "Run this command in order to renew Kubernetes cluster certificates",
		Run: func(cmd *cobra.Command, args []string) {
			if opts.configFile == "" {
				fmt.Println("missing '--config' with arguments")
				os.Exit(1)
			}
			if err := opts.run(); err != nil {
				klog.Error(err)
				return
			}
		},
	}

	cmd.Flags().StringVar(&opts.configFile, "config", "", "Using the config file to renew certificates.")

	return cmd
}

func (o *renewOptions) run() error {
	certificatesDir, err := kubeadm.GetCertificatesDirFromConfigFile(o.configFile)
	if err != nil {
		return err
	}
	o.kubernetesDir = filepath.Dir(certificatesDir)

	// 1. backup old certificates to temp dir.
	fmt.Printf("[renew] Backup old Kubernetes certificates directory %s \n", certificatesDir)
	if err := certs.BackupCertificates(certificatesDir, ""); err != nil {
		return err
	}

	// 2. remove old certificates exclude CA and sa.
	fmt.Println("[renew] Remove old Kubernetes certificates exclude CA and sa")
	if err := certs.RemoveOldCertificates(certificatesDir); err != nil {
		return err
	}

	// 3. remove old kubeconfig
	fmt.Println("[renew] Remove old kubeconfig file")
	if err := kubeconfig.RemoveOldKubeconfig(o.kubernetesDir); err != nil {
		return err
	}

	// 4. renew certificates and kubeconfig
	fmt.Println("[renew] Renew Kubernetes certificates")
	if err := certs.RenewCertificate(o.configFile); err != nil {
		return err
	}

	fmt.Println("[renew] Renew Kubernetes components kubeconfig")
	if err := kubeconfig.RenewKubeConfigFile(o.configFile); err != nil {
		return err
	}

	// 5. remove kubelet certificates
	fmt.Println("[renew] Remove old kubelet certificates")
	if err := certs.RemoveKubeletCertificate(constants.KubeletCertificatesPath); err != nil {
		return err
	}

	// 6. copy new admin.conf to $HOME/.kube/config
	fmt.Println("[renew] Copy admin.conf to $HOME/.kube/config")
	if err := kubeconfig.CreateKubectlKubeConfig(o.kubernetesDir); err != nil {
		return err
	}

	return nil
}
