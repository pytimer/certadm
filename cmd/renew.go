package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pytimer/certadm/pkg/certs"
	"github.com/pytimer/certadm/pkg/constants"
	"github.com/pytimer/certadm/pkg/kubeadm"
	"github.com/pytimer/certadm/pkg/kubeconfig"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/initsystem"
	"k8s.io/utils/exec"
)

var controlPlaneNames = []string{
	"kube-apiserver",
	"kube-scheduler",
	"kube-controller-manager",
	"etcd",
}

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

	// 7. restart control-plane components and kubelet service
	// Try to restart control plane components
	// TODO: crictl not implement now, it only support docker.
	// crictlPath, err := exec.LookPath("crictl")
	// if err == nil {
	// 	restartWithCrictl()
	// } else {
	// 	restartWithDocker()
	// }
	if err := resetWithDocker(); err != nil {
		klog.Errorln("[renew] failed to stop the running control plane containers")
		klog.Warningln("[renew] please stop the running control plane containers manually")
	}

	fmt.Printf("[renew] waiting for the kubelet to boot up the control plane as Static Pods from %s/manifests \n", o.kubernetesDir)
	if err := waitForContainersRunning(); err != nil {
		klog.Errorf("[renew] failed to waiting for containers running: [%v]\n", err)
		klog.Warningln("[renew] please ensure control plane running by docker or crictl")
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
			klog.Warningf("[renew] the kubelet service could not be restarted by kubeadm: [%v]\n", err)
			klog.Warningln("[renew] please ensure kubelet is restarted manually")
		}
	}

	return nil
}

func resetWithDocker() error {
	var filterQuery string
	for _, n := range controlPlaneNames {
		filterQuery += fmt.Sprintf("--filter name=k8s_%s ", n)
	}

	cmd := fmt.Sprintf("docker ps -a %s -q | xargs -r docker rm --force --volumes", filterQuery)
	klog.V(1).Infof("stop the control plane containers, command: [%s]", cmd)
	c := exec.New().Command("sh", "-c", cmd)
	b, err := c.Output()
	klog.Infoln(string(b))
	if err != nil {
		return err
	}
	return nil
}

func waitForContainersRunning() error {
	return wait.PollImmediate(constants.ContainerCallRetryInterval, constants.ContainerCallTimeout, func() (done bool, err error) {
		var errMsg string
		for _, n := range controlPlaneNames {
			cmd := fmt.Sprintf("docker ps -a --filter status=running --filter name=k8s_%s --format '{{ .Names }}'", n)
			klog.V(2).Infof("get %s container status, command: [%s]", n, cmd)
			c := exec.New().Command("sh", "-c", cmd)
			b, err := c.Output()
			if err != nil {
				errMsg += fmt.Sprintf("failed to get %s container status, %s", n, string(b))
				continue
			}

			if string(b) == "" || strings.EqualFold(string(b), "\n") {
				klog.V(3).Infof("container [%s] status not running \n", n)
				errMsg += fmt.Sprintf("%s container not running. ", n)
				continue
			}
		}
		if errMsg != "" {
			klog.V(1).Infof("failed to waiting for the containers running status, %s", errMsg)
			return false, nil
		}
		return true, nil
	})
}
