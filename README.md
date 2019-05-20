# Certadm

Certadm is a tool that renew or recreate Kubernetes cluster certificates created by kubeadm.

Kubernetes cluster created by kubeadm, the certificates expire time is one year, after expired time, cluster can't working. So we should manually renew certificates and make it effect.

This tool support two commands to renew/recreate Kubernetes cluster certificates, it's scope is limited to the local node filesystem now, but it can be integrated with higher level tools.

## Certadm commands

**certadm renew --config=xx.yaml** to renew Kubernetes control-plane components certificates.

**certadm recreate --config=xx.yaml** to recreate Kubernetes control-plane and kubelet components certificates. (This command not implement currently)

## Implement workflow

When the Kubernetes cluster change ip address, the **recreate** command should be used. But this operate is dangerous, so run this command must be carefully.

### Renew command workflow

1. backup old certificates

2. remove old certificates exclude CA and sa, the default certificates directory `/etc/kubernetes/pki`.

`find /etc/kuberentes/pki/ -type f ! -name "ca.*" ! -name "sa.*" ! -name "authwebhookconfig.yaml" | xargs rm`

3. remove control-plane components kubeconfig.

`rm /etc/kubernetes/*.conf`

4. create the new certificates according to config file. It will invoke kubeadm command.

`kubeadm alpha phase certs all  --config=xx.yaml`

`kubeadm alpha phase kubeconfig all --config=xx.yaml`

5. remove kubelet certificates.

`rm /var/libe/kubelet/pki/*`

6. recreate kubectl default kubeconfig.

`cp /etc/kubernetes/pki/admin.conf ~/.kube/config`

## Development

### build

`make build`

## Examples

### Renew certificates and kubeconfig

- kubeadm-cert.yaml文件内容

```yaml
apiVersion: kubeadm.k8s.io/v1alpha2
kind: MasterConfiguration
kubernetesVersion: v1.11.6
apiServerCertSANs:
- "192.168.10.10"
- "k8s-1"
- "192.168.10.100"
- "cloud.kubernetes.cluster.lb"
- "127.0.0.1"
api:
    controlPlaneEndpoint: "cloud.kubernetes.cluster.lb:6443"
etcd:
    local:
        serverCertSANs:
        - "192.168.10.10"
        - "192.168.10.100"
        - "cloud.kubernetes.cluster.lb"
        - "127.0.0.1"
        peerCertSANs:
        - "192.168.10.10"
certificatesDir: /etc/kubernetes/pki
```

```bash
$ ./bin/certadm renew --config=kubeadm-cert.yaml -v 5
[renew] Backup old Kubernetes certificates
I0521 03:01:51.840036   15773 certs.go:38] [certs] Backup certificates from /etc/kubernetes/pki to /tmp/certadm-199797690
[renew] Remove old Kubernetes certificates exclude CA and sa
[renew] Remove old kubeconfig file
[renew] Renew Kubernetes certificates
I0521 03:01:51.847127   15773 kubeadm.go:50] get kubeadm version
I0521 03:01:51.967841   15773 kubeadm.go:62] [kubeadm-certs] Kubeadm version is v1.11.6
I0521 03:01:51.968069   15773 kubeadm.go:67] [kubeadm-certs] renew certificates command args: '/bin/kubeadm alpha phase certs all --config=kubeadm-cert.yaml'
I0521 03:01:58.380451   15773 certs.go:79] [endpoint] WARNING: port specified in api.controlPlaneEndpoint overrides api.bindPort in the controlplane address
[certificates] Using the existing ca certificate and key.
[certificates] Using the existing apiserver certificate and key.
[certificates] Using the existing apiserver-kubelet-client certificate and key.
[certificates] Using the existing sa key.
[certificates] Using the existing front-proxy-ca certificate and key.
[certificates] Using the existing front-proxy-client certificate and key.
[certificates] Using the existing etcd/ca certificate and key.
[certificates] Using the existing etcd/server certificate and key.
[certificates] Using the existing etcd/peer certificate and key.
[certificates] Using the existing etcd/healthcheck-client certificate and key.
[certificates] Using the existing apiserver-etcd-client certificate and key.
[certificates] valid certificates and keys now exist in "/etc/kubernetes/pki"
[renew] Renew Kubernetes components kubeconfig
I0521 03:01:58.380597   15773 kubeadm.go:50] get kubeadm version
I0521 03:01:58.480582   15773 kubeadm.go:74] [kubeadm-kubeconfig] Kubeadm version is v1.11.6
I0521 03:01:58.480798   15773 kubeadm.go:79] [kubeadm-kubeconfig] renew kubeconfig command args: '/bin/kubeadm alpha phase kubeconfig all --config=kubeadm-cert.yaml'
I0521 03:02:00.136466   15773 kubeconfig.go:44] [endpoint] WARNING: port specified in api.controlPlaneEndpoint overrides api.bindPort in the controlplane address
[endpoint] WARNING: port specified in api.controlPlaneEndpoint overrides api.bindPort in the controlplane address
[kubeconfig] Wrote KubeConfig file to disk: "/etc/kubernetes/admin.conf"
[kubeconfig] Wrote KubeConfig file to disk: "/etc/kubernetes/kubelet.conf"
[kubeconfig] Wrote KubeConfig file to disk: "/etc/kubernetes/controller-manager.conf"
[kubeconfig] Wrote KubeConfig file to disk: "/etc/kubernetes/scheduler.conf"
[renew] Remove old kubelet certificates
[renew] Copy admin.conf to $HOME/.kube/config
```
