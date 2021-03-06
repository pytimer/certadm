package util

import (
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"

	"github.com/pytimer/certadm/pkg/constants"

	"github.com/pkg/errors"
	errorsutil "k8s.io/apimachinery/pkg/util/errors"
	utilsexec "k8s.io/utils/exec"
)

// ContainerRuntime is an interface for working with container runtimes
type ContainerRuntime interface {
	IsDocker() bool
	IsRunning() error
	ListKubeContainers() ([]string, error)
	RemoveContainers(containers []string) error
	PullImage(image string) error
	ImageExists(image string) (bool, error)
}

// CRIRuntime is a struct that interfaces with the CRI
type CRIRuntime struct {
	exec      utilsexec.Interface
	criSocket string
}

// DockerRuntime is a struct that interfaces with the Docker daemon
type DockerRuntime struct {
	exec utilsexec.Interface
}

// NewContainerRuntime sets up and returns a ContainerRuntime struct
func NewContainerRuntime(execer utilsexec.Interface, criSocket string) (ContainerRuntime, error) {
	var toolName string
	var runtime ContainerRuntime

	if criSocket != constants.DefaultDockerCRISocket {
		toolName = "crictl"
		// !!! temporary work around crictl warning:
		// Using "/var/run/crio/crio.sock" as endpoint is deprecated,
		// please consider using full url format "unix:///var/run/crio/crio.sock"
		if filepath.IsAbs(criSocket) && goruntime.GOOS != "windows" {
			criSocket = "unix://" + criSocket
		}
		runtime = &CRIRuntime{execer, criSocket}
	} else {
		toolName = "docker"
		runtime = &DockerRuntime{execer}
	}

	if _, err := execer.LookPath(toolName); err != nil {
		return nil, errors.Wrapf(err, "%s is required for container runtime", toolName)
	}

	return runtime, nil
}

// IsDocker returns true if the runtime is docker
func (runtime *CRIRuntime) IsDocker() bool {
	return false
}

// IsDocker returns true if the runtime is docker
func (runtime *DockerRuntime) IsDocker() bool {
	return true
}

// IsRunning checks if runtime is running
func (runtime *CRIRuntime) IsRunning() error {
	if out, err := runtime.exec.Command("crictl", "-r", runtime.criSocket, "info").CombinedOutput(); err != nil {
		return errors.Wrapf(err, "container runtime is not running: output: %s, error", string(out))
	}
	return nil
}

// IsRunning checks if runtime is running
func (runtime *DockerRuntime) IsRunning() error {
	if out, err := runtime.exec.Command("docker", "info").CombinedOutput(); err != nil {
		return errors.Wrapf(err, "container runtime is not running: output: %s, error", string(out))
	}
	return nil
}

// ListKubeContainers lists running k8s CRI pods
func (runtime *CRIRuntime) ListKubeContainers() ([]string, error) {
	out, err := runtime.exec.Command("crictl", "-r", runtime.criSocket, "pods", "--label", "tier=control-plane", "-q").CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "output: %s, error", string(out))
	}
	pods := []string{}
	for _, pod := range strings.Fields(string(out)) {
		pods = append(pods, pod)
	}
	return pods, nil
}

// ListKubeContainers lists running k8s containers
func (runtime *DockerRuntime) ListKubeContainers() ([]string, error) {
	filterQuery := []string{"ps", "-a", "-q"}
	for _, n := range constants.ControlPlaneNames {
		filterQuery = append(filterQuery, "--filter")
		filterQuery = append(filterQuery, fmt.Sprintf("name=k8s_%s", n))
	}
	output, err := runtime.exec.Command("docker", filterQuery...).CombinedOutput()
	return strings.Fields(string(output)), err
}

// RemoveContainers removes running k8s pods
func (runtime *CRIRuntime) RemoveContainers(containers []string) error {
	errs := []error{}
	for _, container := range containers {
		out, err := runtime.exec.Command("crictl", "-r", runtime.criSocket, "stopp", container).CombinedOutput()
		if err != nil {
			// don't stop on errors, try to remove as many containers as possible
			errs = append(errs, errors.Wrapf(err, "failed to stop running pod %s: output: %s, error", container, string(out)))
		} else {
			out, err = runtime.exec.Command("crictl", "-r", runtime.criSocket, "rmp", container).CombinedOutput()
			if err != nil {
				errs = append(errs, errors.Wrapf(err, "failed to remove running container %s: output: %s, error", container, string(out)))
			}
		}
	}
	return errorsutil.NewAggregate(errs)
}

// RemoveContainers removes running containers
func (runtime *DockerRuntime) RemoveContainers(containers []string) error {
	errs := []error{}
	for _, container := range containers {
		out, err := runtime.exec.Command("docker", "rm", "--force", "--volumes", container).CombinedOutput()
		if err != nil {
			// don't stop on errors, try to remove as many containers as possible
			errs = append(errs, errors.Wrapf(err, "failed to remove running container %s: output: %s, error", container, string(out)))
		}
	}
	return errorsutil.NewAggregate(errs)
}

// PullImage pulls the image
func (runtime *CRIRuntime) PullImage(image string) error {
	out, err := runtime.exec.Command("crictl", "-r", runtime.criSocket, "pull", image).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "output: %s, error", string(out))
	}
	return nil
}

// PullImage pulls the image
func (runtime *DockerRuntime) PullImage(image string) error {
	out, err := runtime.exec.Command("docker", "pull", image).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "output: %s, error", string(out))
	}
	return nil
}

// ImageExists checks to see if the image exists on the system
func (runtime *CRIRuntime) ImageExists(image string) (bool, error) {
	err := runtime.exec.Command("crictl", "-r", runtime.criSocket, "inspecti", image).Run()
	return err == nil, nil
}

// ImageExists checks to see if the image exists on the system
func (runtime *DockerRuntime) ImageExists(image string) (bool, error) {
	err := runtime.exec.Command("docker", "inspect", image).Run()
	return err == nil, nil
}

// isExistingSocket checks if path exists and is domain socket
func isExistingSocket(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.Mode()&os.ModeSocket != 0
}

// detectCRISocketImpl is separated out only for test purposes, DON'T call it directly, use DetectCRISocket instead
func detectCRISocketImpl(isSocket func(string) bool) (string, error) {
	const (
		dockerSocket     = "/var/run/docker.sock" // The Docker socket is not CRI compatible
		containerdSocket = "/run/containerd/containerd.sock"
	)

	foundCRISockets := []string{}
	knownCRISockets := []string{
		// Docker and containerd sockets are special cased below, hence not to be included here
		"/var/run/crio/crio.sock",
	}

	if isSocket(dockerSocket) {
		// the path in dockerSocket is not CRI compatible, hence we should replace it with a CRI compatible socket
		foundCRISockets = append(foundCRISockets, constants.DefaultDockerCRISocket)
	} else if isSocket(containerdSocket) {
		// Docker 18.09 gets bundled together with containerd, thus having both dockerSocket and containerdSocket present.
		// For compatibility reasons, we use the containerd socket only if Docker is not detected.
		foundCRISockets = append(foundCRISockets, containerdSocket)
	}

	for _, socket := range knownCRISockets {
		if isSocket(socket) {
			foundCRISockets = append(foundCRISockets, socket)
		}
	}

	switch len(foundCRISockets) {
	case 0:
		// Fall back to Docker if no CRI is detected, we can error out later on if we need it
		return constants.DefaultDockerCRISocket, nil
	case 1:
		// Precisely one CRI found, use that
		return foundCRISockets[0], nil
	default:
		// Multiple CRIs installed?
		return "", errors.Errorf("Found multiple CRI sockets, please use --cri-socket to select one: %s", strings.Join(foundCRISockets, ", "))
	}
}

// DetectCRISocket uses a list of known CRI sockets to detect one. If more than one or none is discovered, an error is returned.
func DetectCRISocket() (string, error) {
	if goruntime.GOOS != "linux" {
		return constants.DefaultDockerCRISocket, nil
	}

	return detectCRISocketImpl(isExistingSocket)
}
