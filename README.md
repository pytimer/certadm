# Certadm

Certadm is a tool that renew Kubernetes cluster certificates created by kubeadm.

Kubernetes cluster created by kubeadm, the certificates expire time is one year, after expired time, cluster can't working. So we should manually renew certificates and make it effect.

This tool support renew Kubernetes cluster certificates, it's scope is limited to the local node filesystem now, but it can be integrated with higher level tools.

## Support Kubeadm version

Now the tool only support v1.11.0+, and the latest kubeadm support renew command, so we can use `kubeadm renew` to renew certificates, the latest kubeadm version will be support.

If the Kubernetes version v1.13.0+, you can see [renew certficates](https://github.com/kubernetes/kubeadm/issues/581#issuecomment-471575078) .

## Certadm commands

**certadm renew --config=xx.yaml** to renew Kubernetes control-plane components certificates.

## Implement workflow

### Renew command workflow

1. backup old certificates

2. remove old certificates exclude CA and sa, the default certificates directory `/etc/kubernetes/pki`.

`find /etc/kuberentes/pki/ -type f ! -name "ca.*" ! -name "sa.*" | xargs rm`

3. remove control-plane components kubeconfig.

`rm /etc/kubernetes/*.conf`

4. create the new certificates according to config file. It will invoke kubeadm command.

`kubeadm alpha phase certs all  --config=xx.yaml`

`kubeadm alpha phase kubeconfig all --config=xx.yaml`

5. remove kubelet certificates.

`rm /var/libe/kubelet/pki/*`

6. recreate kubectl default kubeconfig.

`cp /etc/kubernetes/pki/admin.conf ~/.kube/config`

7. restart control plane containers and kubelet service

## Development

### build

`make build`

## Examples

### Renew certificates and kubeconfig

- kubeadm-cert.yaml content:

```yaml
apiVersion: kubeadm.k8s.io/v1alpha2
kind: MasterConfiguration
kubernetesVersion: v1.11.6
apiServerCertSANs:
- "192.168.10.10"
- "k8s-1"
- "192.168.10.100"
- "cloud.kubernetes.cluster.lb"
api:
    controlPlaneEndpoint: "cloud.kubernetes.cluster.lb:6443"
etcd:
    local:
        serverCertSANs:
        - "192.168.10.10"
        - "192.168.10.100"
        - "cloud.kubernetes.cluster.lb"
        peerCertSANs:
        - "192.168.10.10"
certificatesDir: /etc/kubernetes/pki
nodeRegistration:
    name: "k8s-1"
```

Renew the Kubernetes certificates.

```bash
$ ./bin/certadm renew --config=kubeadm-cert.yaml
I0529 18:24:03.660050    5059 renew.go:53] [renew] Detected and using CRI socket: /var/run/dockershim.sock
[renew] Backup old Kubernetes certificates directory /etc/kubernetes/pki
[renew] Remove old Kubernetes certificates exclude CA and sa
[renew] Remove old kubeconfig file
[renew] Renew Kubernetes certificates
I0529 18:24:10.196249    5059 certs.go:81] [endpoint] WARNING: port specified in api.controlPlaneEndpoint overrides api.bindPort in the controlplane address
[certificates] Using the existing ca certificate and key.
[certificates] Generated apiserver certificate and key.
[certificates] apiserver serving cert is signed for DNS names [k8s-1 kubernetes kubernetes.default kubernetes.default.svc kubernetes.default.svc.cluster.local cloud.kubernetes.cluster.lb k8s-1 cloud.kubernetes.cluster.lb] and IPs [10.96.0.1 192.168.10.10 192.168.10.10 192.168.10.100]
[certificates] Generated apiserver-kubelet-client certificate and key.
[certificates] Using the existing sa key.
[certificates] Generated front-proxy-ca certificate and key.
[certificates] Generated front-proxy-client certificate and key.
[certificates] Using the existing etcd/ca certificate and key.
[certificates] Generated etcd/server certificate and key.
[certificates] etcd/server serving cert is signed for DNS names [k8s-1 localhost cloud.kubernetes.cluster.lb] and IPs [127.0.0.1 ::1 192.168.10.10 192.168.10.100]
[certificates] Generated etcd/peer certificate and key.
[certificates] etcd/peer serving cert is signed for DNS names [k8s-1 localhost] and IPs [192.168.10.10 127.0.0.1 ::1 192.168.10.10]
[certificates] Generated etcd/healthcheck-client certificate and key.
[certificates] Generated apiserver-etcd-client certificate and key.
[certificates] valid certificates and keys now exist in "/etc/kubernetes/pki"
[renew] Renew Kubernetes components kubeconfig
I0529 18:24:11.283982    5059 kubeconfig.go:45] [endpoint] WARNING: port specified in api.controlPlaneEndpoint overrides api.bindPort in the controlplane address
[endpoint] WARNING: port specified in api.controlPlaneEndpoint overrides api.bindPort in the controlplane address
[kubeconfig] Wrote KubeConfig file to disk: "/etc/kubernetes/admin.conf"
[kubeconfig] Wrote KubeConfig file to disk: "/etc/kubernetes/kubelet.conf"
[kubeconfig] Wrote KubeConfig file to disk: "/etc/kubernetes/controller-manager.conf"
[kubeconfig] Wrote KubeConfig file to disk: "/etc/kubernetes/scheduler.conf"
[renew] Remove old kubelet certificates
[renew] Copy admin.conf to $HOME/.kube/config
I0529 18:24:11.353448    5059 renew.go:168] kubernetes-manager containers: [e3694b0955bd a510bbae0087 fa9ea420be81 951b25ad0cbc]
[renew] waiting for the kubelet to boot up the control plane as Static Pods from /etc/kubernetes/manifests
I0529 18:24:23.811292    5059 renew.go:125] [renew] kubernetes-manager containers running
[renew] restarting the kubelet service
[renew] ensure the kubelet service is active
```
