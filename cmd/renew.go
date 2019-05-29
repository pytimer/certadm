package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pytimer/certadm/pkg/certs"
	"github.com/pytimer/certadm/pkg/constants"
	"github.com/pytimer/certadm/pkg/kubeadm"
	"github.com/pytimer/certadm/pkg/kubeconfig"
	"github.com/pytimer/certadm/pkg/util"
	utilruntime "github.com/pytimer/certadm/pkg/util/runtime"

	"github.com/spf13/cobra"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/initsystem"
	utilsexec "k8s.io/utils/exec"
)

type renewOptions struct {
	kubernetesDir string
	configFile    string
	criSocketPath string
}

// NewCmdRenew returns "certadm renew" command.
func NewCmdRenew() *cobra.Command {
	opts := &renewOptions{}
	cmd := &cobra.Command{
		Use:   "renew",
		Short: "Run this command in order to renew Kubernetes cluster certificates",
		Run: func(cmd *cobra.Command, args []string) {

			if opts.configFile != "" {
				c, err := kubeadm.FetchConfigurationFromConfigFile(opts.configFile)
				if err != nil {
					klog.Errorf("failed to load config from '--config': [%v]", err)
					os.Exit(1)
				}

				if c.CertificatesDir == "" {
					klog.Warningf("missing the Kubernetes root directory with '--config', so using the default directory %q\n", constants.KubernetesDir)
					c.CertificatesDir = filepath.Join(constants.KubernetesDir, "pki")
				}
				opts.kubernetesDir = filepath.Dir(c.CertificatesDir)

				opts.criSocketPath, err = DetectCRISocket(c)
				if err != nil {
					klog.Warningf("[renew] failed to detected and using CRI socket: %v", err)
					opts.criSocketPath = constants.DefaultDockerCRISocket
				}
				klog.Infof("[renew] Detected and using CRI socket: %s", opts.criSocketPath)
			}

			if err := opts.run(); err != nil {
				klog.Error(err)
				return
			}
		},
	}

	cmd.Flags().StringVar(&opts.configFile, "config", "", "Using the config file to renew certificates.")
	cmd.Flags().StringVar(&opts.kubernetesDir, "root-dir", constants.KubernetesDir, "The path save the Kubernetes certificates and kubeconfig.")

	return cmd
}

func (o *renewOptions) run() error {
	certificatesDir := filepath.Join(o.kubernetesDir, "pki")

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

	// 7. restart control-plane components and kubelet service
	// Try to restart control plane components
	if err := removeContainers(utilsexec.New(), o.criSocketPath); err != nil {
		klog.Errorf("[renew] failed to stop the control plane containers, %v \n", err)
		klog.Warningln("[renew] please stop the control plane containers manually")
	}

	fmt.Printf("[renew] waiting for the kubelet to boot up the control plane as Static Pods from %s/manifests \n", o.kubernetesDir)
	if err := util.WaitForContainersRunning(constants.ControlPlaneNames); err != nil {
		klog.Errorf("[renew] failed to waiting for containers running: [%v]\n", err)
		klog.Warningln("[renew] please ensure control plane running by docker or crictl")
	}else {
		klog.Infoln("[renew] kubernetes-manager containers running")
	}

	// Try to restart the kubelet service
	klog.V(1).Infoln("[renew] getting init system")
	initSystem, err := initsystem.GetInitSystem()
	if err != nil {
		klog.Warningln("[renew] the kubelet service could not restarted by certadm. Unable to detect a supported init system!")
		klog.Warningln("[renew] please ensure kubelet is restarted manually")
	} else {
		fmt.Println("[renew] restarting the kubelet service")
		if err := initSystem.ServiceRestart("kubelet"); err != nil {
			klog.Warningf("[renew] the kubelet service could not be restarted by certadm: [%v]\n", err)
			klog.Warningln("[renew] please ensure kubelet is restarted manually")
		}

		fmt.Println("[renew] ensure the kubelet service is active")
		if err := util.WaitForServiceActive("kubelet", constants.ServiceCallRetryInterval, constants.ServiceCallTimeout); err != nil {
			klog.Warningln("[wait-service] please ensure kubelet is active manually")
		}
	}

	return nil
}

func DetectCRISocket(cfg *kubeadm.Config) (string, error) {
	if cfg != nil && cfg.CRISocket != "" {
		return cfg.CRISocket, nil
	}

	return utilruntime.DetectCRISocket()
}

func removeContainers(execer utilsexec.Interface, criSocketPath string) error {
	containerRuntime, err := utilruntime.NewContainerRuntime(execer, criSocketPath)
	if err != nil {
		return err
	}
	klog.V(1).Infof("container runtime %v", containerRuntime)
	containers, err := containerRuntime.ListKubeContainers()
	if err != nil {
		return err
	}
	klog.Infof("kubernetes-manager containers: %v", containers)
	return containerRuntime.RemoveContainers(containers)
}
