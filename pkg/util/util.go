package util

import (
	"fmt"
	"strings"
	"time"

	"github.com/pytimer/certadm/pkg/constants"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/util/initsystem"
	"k8s.io/utils/exec"
)

func WaitForContainersRunning(containers []string) error {
	return wait.PollImmediate(constants.ContainerCallRetryInterval, constants.ContainerCallTimeout, func() (done bool, err error) {
		var errMsg string
		for _, n := range containers {
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
			klog.V(1).Infof("failed to waiting for the containers running status, %s, retry...", errMsg)
			return false, nil
		}
		return true, nil
	})
}

func WaitForServiceActive(name string, interval, timeout time.Duration) error {
	return wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		initSystem, err := initsystem.GetInitSystem()
		if err != nil {
			klog.Warningf("[wait-service] the %s service could not started by certadm. Unable to detect a supported init system!\n", name)
			return false, nil
		}

		if initSystem.ServiceIsActive(name) {
			return true, nil
		}

		return false, nil
	})
}
