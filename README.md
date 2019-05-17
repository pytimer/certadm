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


 