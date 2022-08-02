# 使用 kubeadm 进行证书管理
**由于环境使用的 kubeadm 部署，所以只对 kubeadm 进行证书管理分析。**
**文档只用于记录学习[详细信息可以查看官网](https://kubernetes.io/zh-cn/docs/tasks/administer-cluster/kubeadm/kubeadm-certs/)**

由 kubeadm 生成的客户端证书在 1 年后到期，我们需要分析证书到期续签策略以及证书的影响。
### 检查证书

` 你可以使用 check-expiration 子命令来检查证书何时过期 `

```
[root@dev01 ~]# kubeadm certs check-expiration

CERTIFICATE                EXPIRES                  RESIDUAL TIME   CERTIFICATE AUTHORITY   EXTERNALLY MANAGED
admin.conf                 May 24, 2023 03:16 UTC   320d                                    no
apiserver                  May 24, 2023 03:16 UTC   320d            ca                      no
apiserver-kubelet-client   May 24, 2023 03:16 UTC   320d            ca                      no
controller-manager.conf    May 24, 2023 03:16 UTC   320d                                    no
front-proxy-client         May 24, 2023 03:16 UTC   320d            front-proxy-ca          no
scheduler.conf             May 24, 2023 03:16 UTC   320d                                    no

CERTIFICATE AUTHORITY   EXPIRES                  RESIDUAL TIME   EXTERNALLY MANAGED
ca                      May 21, 2032 03:16 UTC   9y              no
front-proxy-ca          May 21, 2032 03:16 UTC   9y              no
```
`
通过 kubeadm certs check-expiration 可以发现 kubeadm 帮我们查询了
`

* 用户账户证书：admin.conf、scheduler.conf、controller-manager.conf
* 服务证书：apiserver、apiserver-kubelet-client、front-proxy-client
* 签发机构证书：ca、front-proxy-ca

**此处发现没有印象中的 etcd 证书与 kubelet 证书。**

`kubeadm 查询证书是根据某些路径去查询，默认 /etc/kubernetes/pki 查询服务证书文件，/etc/kubernetes 查询用户账户证书。
`
`
如果你新增一个 example.conf 或者 example.crt，kubeadm是不会查询的，kubeadm 只会查询默认证书。
`
### 各个证书作用
**Kubernetes 需要 PKI 证书才能进行基于 TLS 的身份验证。**

`查看一下环境有哪些证书文件`
```
[root@dev01 pki]# tree
.
├── apiserver.crt
├── apiserver.key
├── apiserver-kubelet-client.crt
├── apiserver-kubelet-client.key
├── ca.crt
├── ca.key
├── front-proxy-ca.crt
├── front-proxy-ca.key
├── front-proxy-client.crt
├── front-proxy-client.key
├── sa.key
└── sa.pub
```

**常用的证书体系需要用到：颁发机构证书、私钥、服务证书文件。**

`目录中 ca.crt 就是颁发机构证书所有服务证书都是依靠此证书生成，apiserver.key 就是私钥，apiserver.crt 就是服务证书文件，k8s 中颁发机构是 10 年，私钥没有时间，所以就是要保障服务证书文件的时间不要过期。`

**各个证书的介绍**
```
apiserver.crt  用于对外提供服务的 apiserver 证书及私钥
apiserver.key 
```
```
apiserver-kubelet-client.crt  用于 apiserver 访问 kubelet 所需的客户端证书及私钥
apiserver-kubelet-client.key
```
```
ca.crt  用于签发 k8s 中其他证书的颁发机构证书及私钥
ca.key
```
```
front-proxy-ca.crt  用于签发 k8s 中前端代理的颁发机构证书及私钥
front-proxy-ca.key
front-proxy-client.crt  用于 apiserver 扩展的前端代理所需的客户端证书及私钥
front-proxy-client.key
```
```
sa.key  验证 serviceaccount token 的公钥
sa.pub
提供给 kube-controller-manager使用，kube-controller-manager 创建pod时通过 sa.key 对 token 进行签名，apiserver通过公钥 sa.pub 进行签名的验证 如 kube-proxy 是以 pod 形式运行的, 在 pod 中，直接使用 service account 与 kube-apiserver 进行认证，此时就不需要再单独为 kube-proxy 创建证书了，会直接使用token校验。
```
`大致了解证书文件的作用后，我们发现目录下没有 scheduler、controller-manager 证书文件`
`这时由于controller-manager、scheduler、kubelet、kube-proxy等组件使用的是用户账户证书与 apiserver 通信，所以此时 /etc/kubernetes/ 下的 *.conf 文件发挥了作用`
```
admin.conf  集群管理员的客户端证书，用于 API 服务器身份认证
```
```
scheduler.conf  调度器的客户端证书，用于和 apiserver 的会话
```
```
controller-manager.conf 控制器的客户端证书，用于和 apiserver 服务器的会话
```
`查看 kube-controller-manager.yaml，发现里面没有像 apiserver 一样使用各种 crt 证书文件，而是在认证的地方使用的 conf，这说明 conf 文件包含了证书颁发机构、证书文件、私钥`
```
[root@dev01 manifests]#  cat kube-controller-manager.yaml
apiVersion: v1
kind: Pod
metadata:
  ...
  name: kube-controller-manager
  namespace: kube-system
spec:
  containers:
  - command:
    - kube-controller-manager
    - --authentication-kubeconfig=/etc/kubernetes/controller-manager.conf
    - --authorization-kubeconfig=/etc/kubernetes/controller-manager.conf
    - --client-ca-file=/etc/kubernetes/ssl/ca.crt
    ...
```
**查看 controller-manager.conf 文件格式**
```
[root@dev01 kubernetes]# cat controller-manager.conf
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM1ekNDQWMrZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeU1EVXlOREF6TVRZek1Wb1hEVE15TURVeU1UQXpNVFl6TVZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBUER4CnFjQitOZlJjY3hXY1l0N2ltREVLTVN2UnU5RUQ3K3Zha1pMeURuL0JKQmNQSXNRVDdnVHhhZkwxQVZlMWY4dU4KY09wRE9uNmgwUWxHVzI4WXpPYjBUSnJ6TEs1Q2ZHOWpvS0p0emxXdkJrU2tiNTV5enYySWlWTXJMNW9pbmJ1UAo0dGVKeCtmNWFKQlRiRFBDN2hjSTRaTDNPUXAwc0Q5OXZTbFNOeVcxTi9tZDFmRjJIY2dmcEdaZjNhUEFnSW1yCnp6TFJka2xDNVdHcGxRZFNzWlhheUxoYUIzOXFPdDZVQ2NCSVI4R0lVKzhQZDdkNGpOUEpVUW5oUFNMakI0cXYKNzN2YkVyL2dOM05vQ1pXN0ZyMFBYUk44dGRSTDFiNVBPS2JMUHdyODhDYlI4QjNqdGQ1aTB0RC9DdDNRVUhragppZ1YwdDNGNTlqZFBPUWN4NUUwQ0F3RUFBYU5DTUVBd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZDbFc2cTMrcm5MNnh5TFFSZWxHNVBzYmpmaFRNQTBHQ1NxR1NJYjMKRFFFQkN3VUFBNElCQVFCV0htbmpYL0hpS3Z5VnBPSlU3YzR2ZElnbFdWS1lKellCV1dKa0Rhb3RzMlViUG1PSgpieElwYkh5UW9Hc25WaVBEc1NTUGxJZW9nWStTc3BDT2N3ZVVYanFLWWNpS3VaSFRQRGlLYlhqWTVyelRmcnAwCkRFNlJVMGdKTU1BQTlJd1hreFpiNTBxaTlYRDlPT0JjM3VRaENFQWRkenIyRHFPZ29aZENJdk5KbkQ4Zk9XK0gKbXBKRWhPM1kwclBHWFV5K2JQOHZVa2h2ZktYVFFDem90MjNLaCtmRzZvVU5rQkc0UXowaTBMV1JLa1hXdTE0bgpTa0Nqc1FuZzBVMUJvUzFVWm42WXpwSHJZbm5WTDdOeVFwTWJCeGNTaEt3eU96ZVRjSHl5VUJVVDE2T0thODdHClhBR0JKYVRodmdiR0U0NTZEd0VwUmtzU3ZDZXk5MWJDN3FCRAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    server: https://127.0.0.1:6443
  name: cluster.local
contexts:
- context:
    cluster: cluster.local
    user: system:kube-controller-manager
  name: system:kube-controller-manager@cluster.local
current-context: system:kube-controller-manager@cluster.local
kind: Config
preferences: {}
users:
- name: system:kube-controller-manager
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURDRENDQWZDZ0F3SUJBZ0lJVWtyU2lrUENLSnN3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TWpBMU1qUXdNekUyTXpGYUZ3MHlNekExTWpRd016RTJNek5hTUNreApKekFsQmdOVkJBTVRIbk41YzNSbGJUcHJkV0psTFdOdmJuUnliMnhzWlhJdGJXRnVZV2RsY2pDQ0FTSXdEUVlKCktvWklodmNOQVFFQkJRQURnZ0VQQURDQ0FRb0NnZ0VCQU5iUmUvblVPNFhrenhmZDcwdzhFTXdCcmFxUTJ5NGgKcUJLZE9wc2wra1kxaFpoWFhMY3BiSEU4OXNJOWRLRnZwNmRBVHVrL3dTMFZNRjh3VTZ3QlJ0TkdDYXk4VWxKaApFMTVZQ0d0UzMrMzZrNDMwSGpLS2p0QmlEb1pwTnZWcW5Nc1ptc0x3NzF3Y0RpdTU4QitCNklLeEgrMUNnQlVDCnh6dmVrMDRTMzJEYndsMEMvK1g5Yzl1L1dJZ3VsZkx1Q2wxWHVnWEs4VEdnWkdKVjlIQ1I3Tk9nY1lhN0RCSDcKUm1qakxjVVMxbXJEYzl6dy9LOGtaYVpMUjBYUDVKT2xGN1pVdkhjMTF6cm9yaWpFeHZHVmdsa2pPeEhxVW9aZgpqMU8vTUlVb1JIRHlGeUdoaXdhcW1jUVFPMnQ1RW80QWVKTHU4d3NENUZVYjFUQ3gveldzb3gwQ0F3RUFBYU5JCk1FWXdEZ1lEVlIwUEFRSC9CQVFEQWdXZ01CTUdBMVVkSlFRTU1Bb0dDQ3NHQVFVRkJ3TUNNQjhHQTFVZEl3UVkKTUJhQUZDbFc2cTMrcm5MNnh5TFFSZWxHNVBzYmpmaFRNQTBHQ1NxR1NJYjNEUUVCQ3dVQUE0SUJBUUJRUUxleQpXT1FkQzBrUWw3cGtFa3ZMdTNoN09NeDZSSGdRUW9acVFYU1hIUUkvV2Y1eXNrNWZnV2V0OURzNGhCVGZTamUyClhXQm9IMVVuQTRVbWNPYlhHdmx0RjBKQUNSbTc5bit0d1dmNkJKeTlSRWdwYjVzTHdLb3lQTEZmMkxpQU9nZkMKVFhRVTBCTUpiaFVlY2dxWHN1Z1V1c0M3eFQrWVIxcUUraHNTcjg1aGZ1dStwVkFlWHBBR0ZnRWYwN0t2NVdCTgpOT0dXaFJGWFo3S0p6K2gzSnpZTjdGNkVGNU81cUJ3eW03bGw4RlZYUnVLNDNEVVRuNE1BTExRWm9CbnNXUEQvClNwK1pQUlFpZElCOG15RkFybExiMHZOYU1oUC8wbGdBcHljVUgzYXB5Qkg4WEZ3ZUt2QUgzNFpVNDdCcXF1bjIKK1dsRkNFdjJUTHIxUWZydgotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBMXRGNytkUTdoZVRQRjkzdlREd1F6QUd0cXBEYkxpR29FcDA2bXlYNlJqV0ZtRmRjCnR5bHNjVHoyd2oxMG9XK25wMEJPNlQvQkxSVXdYekJUckFGRzAwWUpyTHhTVW1FVFhsZ0lhMUxmN2ZxVGpmUWUKTW9xTzBHSU9obWsyOVdxY3l4bWF3dkR2WEJ3T0s3bndINEhvZ3JFZjdVS0FGUUxITzk2VFRoTGZZTnZDWFFMLwo1ZjF6Mjc5WWlDNlY4dTRLWFZlNkJjcnhNYUJrWWxYMGNKSHMwNkJ4aHJzTUVmdEdhT010eFJMV2FzTnozUEQ4CnJ5Umxwa3RIUmMva2s2VVh0bFM4ZHpYWE91aXVLTVRHOFpXQ1dTTTdFZXBTaGwrUFU3OHdoU2hFY1BJWElhR0wKQnFxWnhCQTdhM2tTamdCNGt1N3pDd1BrVlJ2Vk1MSC9OYXlqSFFJREFRQUJBb0lCQUI3VXhoUmZMNFBTQ1I4Rgp6VmlGcm10NU9qR214SzVrU2prUnM5MmNuOFQvM1lLMTF5cTZtVVBQV2lNYXF1aVltOWxYVTN3S21IRWQvMmxLCkNGaEJmcWJJQUk5NEgrQ082amE2TDhzbmc5Z21CSk4xRy9ocEJvSkY4SXUrQWhOTEQrSGVkT1hScXFqdnBlRUUKdUwvczV6QUt0K2hYaVFaNjlabFVJcFFzQlQ5alI2VXJpQlNYNDNrSVQvYmJoazRZT001Wlkrc1VjdFlIT0lpYgpvc0JaTjJkZ0xIdjJmUmgzT09JL0xubXZSczQ5YmhDaGhGUnVqaSt4d3ZoYXVrVktWdHRCalZVOEVqV2ViemtRCmsrTTBoSktua1B5RTJ4eHlyYmtEcm9hbHBiWjRCMWR1K2svS1NGcjhKMEJNOG5uZmwzWTR1ZTlicDVvdEx3VE8Kc3U3TDZFRUNnWUVBM3VEbjBSenBHbkVHOG1UUmNya2Z6QzF2QzlnamtIRFhHblN6SE9JNmUySWJacmtJMVhUagpzZWM4T2FwUGdpSmZ3b2tRRm5DUEhIUGx0alVzcjVSYW03a3dTenhLb2RaL3FoZjZNVmpZd2tiZ1RhQkpWZlVuCkVEU2NtZ29KSVY5S25zbUtiQW1uWlI1YVpHYThxNGhVd0NCTXN0U0xZMnFwWElpWUpLUnFzOFVDZ1lFQTlyM3cKUVZnRnRIVnpGQVZVblVhVlIvZzMzNlU3MzllSnNnSnN6cUVmT0dHVkRLYUlOMmQzZEhXczBBdk04MW1qYlFNYwpYbTlHNHBWTnRPN3psb3FSdTh1MHJMUlNSQmtBcE80RGpOYTUwU0dwaTB4WXA1MmpJSGhzQitrOFRhL21JTWFHCjRTK0NyMEJKZ3RvcDVyTlBzUnE0QlJpTEplaHBaeUVkTG1pU3Iza0NnWUVBeWs5Ni9kMXhWdGdjUnpEdTBtVnEKa0tyNUx0WXVTakd6ZjhkSXdzSXo5Y1d6TzhaZ1YycDRPUWUrS2xZSFFBZzF4bDhtSkVoZUQ2VEFjbW5JTzBTKwpEcmJHT0xUc3hnVHY4NVU2UFVyNXp6Z0k2VTNDNCt6WlhhTjBOM3MrbUhWb1JjU0tDdEdraW14SC92VW04OG1YCnJLSmxITHRFSDZVMVFOS1kwRUlTdlVVQ2dZQm5PUlMrc1cxSmptMERMaXUwMUNkMWdWRVduZENMbDJjQ2NtTkUKbWtFdVhCeDN1MTkzRmpEeUpweVNPSzJZREtNQzR1OUdUblV3SzFmajNxZEhEbGRqWWhnR3ZEVkNTZGRzS0NxUwpDKzZXbitlRzMveWl1cDdZNGFnSU16bklJbE02MlBMcVYrN3JGQkRzemdvRGVuWlhXQjh0K3BPVDIwUGxJZ1JNCk92Rkt1UUtCZ1FEU3QvY2ZkYzFhWFd1WjJVcC9qYndUWWoxNjM0TmhuazZlcU5sNVp4Mk1BZEZneS9yeU5LblIKS3RDbWQ0cjZVTmg3cFZKUUdmYzRheHdiblBsR1pxdWg3ektBQlBVVXNxem1EZDFPSGxtSGsybTRtSUhPWkR5Wgo2b3JYN2ZUSDRuS3hJYk9vSG5pWG5NTHI4Q0JhUDFvSytRc0paYTlDdVhkSExBZEw2SnZqUXc9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
```
**发现 conf 文件中有三处密钥对应着证书颁发机构、证书文件、私钥**
`conf 文件中有三处对应证书 certificate-authority-data、client-certificate-data、client-key-data，我们把 certificate-authority-data 通过 base64 转了一下，发现 certificate-authority-data 内容与 ca.crt 一样，说明 controller-manager 还是通过此证书颁发机构颁发了自己的证书与私钥`
```
[root@dev01 kubernetes]# echo "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM1ekNDQWMrZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeU1EVXlOREF6TVRZek1Wb1hEVE15TURVeU1UQXpNVFl6TVZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBUER4CnFjQitOZlJjY3hXY1l0N2ltREVLTVN2UnU5RUQ3K3Zha1pMeURuL0JKQmNQSXNRVDdnVHhhZkwxQVZlMWY4dU4KY09wRE9uNmgwUWxHVzI4WXpPYjBUSnJ6TEs1Q2ZHOWpvS0p0emxXdkJrU2tiNTV5enYySWlWTXJMNW9pbmJ1UAo0dGVKeCtmNWFKQlRiRFBDN2hjSTRaTDNPUXAwc0Q5OXZTbFNOeVcxTi9tZDFmRjJIY2dmcEdaZjNhUEFnSW1yCnp6TFJka2xDNVdHcGxRZFNzWlhheUxoYUIzOXFPdDZVQ2NCSVI4R0lVKzhQZDdkNGpOUEpVUW5oUFNMakI0cXYKNzN2YkVyL2dOM05vQ1pXN0ZyMFBYUk44dGRSTDFiNVBPS2JMUHdyODhDYlI4QjNqdGQ1aTB0RC9DdDNRVUhragppZ1YwdDNGNTlqZFBPUWN4NUUwQ0F3RUFBYU5DTUVBd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZDbFc2cTMrcm5MNnh5TFFSZWxHNVBzYmpmaFRNQTBHQ1NxR1NJYjMKRFFFQkN3VUFBNElCQVFCV0htbmpYL0hpS3Z5VnBPSlU3YzR2ZElnbFdWS1lKellCV1dKa0Rhb3RzMlViUG1PSgpieElwYkh5UW9Hc25WaVBEc1NTUGxJZW9nWStTc3BDT2N3ZVVYanFLWWNpS3VaSFRQRGlLYlhqWTVyelRmcnAwCkRFNlJVMGdKTU1BQTlJd1hreFpiNTBxaTlYRDlPT0JjM3VRaENFQWRkenIyRHFPZ29aZENJdk5KbkQ4Zk9XK0gKbXBKRWhPM1kwclBHWFV5K2JQOHZVa2h2ZktYVFFDem90MjNLaCtmRzZvVU5rQkc0UXowaTBMV1JLa1hXdTE0bgpTa0Nqc1FuZzBVMUJvUzFVWm42WXpwSHJZbm5WTDdOeVFwTWJCeGNTaEt3eU96ZVRjSHl5VUJVVDE2T0thODdHClhBR0JKYVRodmdiR0U0NTZEd0VwUmtzU3ZDZXk5MWJDN3FCRAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==" | base64 -d
-----BEGIN CERTIFICATE-----
MIIC5zCCAc+gAwIBAgIBADANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwprdWJl
cm5ldGVzMB4XDTIyMDUyNDAzMTYzMVoXDTMyMDUyMTAzMTYzMVowFTETMBEGA1UE
AxMKa3ViZXJuZXRlczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAPDx
qcB+NfRccxWcYt7imDEKMSvRu9ED7+vakZLyDn/BJBcPIsQT7gTxafL1AVe1f8uN
cOpDOn6h0QlGW28YzOb0TJrzLK5CfG9joKJtzlWvBkSkb55yzv2IiVMrL5oinbuP
4teJx+f5aJBTbDPC7hcI4ZL3OQp0sD99vSlSNyW1N/md1fF2HcgfpGZf3aPAgImr
zzLRdklC5WGplQdSsZXayLhaB39qOt6UCcBIR8GIU+8Pd7d4jNPJUQnhPSLjB4qv
73vbEr/gN3NoCZW7Fr0PXRN8tdRL1b5POKbLPwr88CbR8B3jtd5i0tD/Ct3QUHkj
igV0t3F59jdPOQcx5E0CAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB
/wQFMAMBAf8wHQYDVR0OBBYEFClW6q3+rnL6xyLQRelG5PsbjfhTMA0GCSqGSIb3
DQEBCwUAA4IBAQBWHmnjX/HiKvyVpOJU7c4vdIglWVKYJzYBWWJkDaots2UbPmOJ
bxIpbHyQoGsnViPDsSSPlIeogY+SspCOcweUXjqKYciKuZHTPDiKbXjY5rzTfrp0
DE6RU0gJMMAA9IwXkxZb50qi9XD9OOBc3uQhCEAddzr2DqOgoZdCIvNJnD8fOW+H
mpJEhO3Y0rPGXUy+bP8vUkhvfKXTQCzot23Kh+fG6oUNkBG4Qz0i0LWRKkXWu14n
SkCjsQng0U1BoS1UZn6YzpHrYnnVL7NyQpMbBxcShKwyOzeTcHyyUBUT16OKa87G
XAGBJaThvgbGE456DwEpRksSvCey91bC7qBD
-----END CERTIFICATE-----
[root@dev01 kubernetes]# cat /etc/kubernetes/pki/ca.crt
-----BEGIN CERTIFICATE-----
MIIC5zCCAc+gAwIBAgIBADANBgkqhkiG9w0BAQsFADAVMRMwEQYDVQQDEwprdWJl
cm5ldGVzMB4XDTIyMDUyNDAzMTYzMVoXDTMyMDUyMTAzMTYzMVowFTETMBEGA1UE
AxMKa3ViZXJuZXRlczCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAPDx
qcB+NfRccxWcYt7imDEKMSvRu9ED7+vakZLyDn/BJBcPIsQT7gTxafL1AVe1f8uN
cOpDOn6h0QlGW28YzOb0TJrzLK5CfG9joKJtzlWvBkSkb55yzv2IiVMrL5oinbuP
4teJx+f5aJBTbDPC7hcI4ZL3OQp0sD99vSlSNyW1N/md1fF2HcgfpGZf3aPAgImr
zzLRdklC5WGplQdSsZXayLhaB39qOt6UCcBIR8GIU+8Pd7d4jNPJUQnhPSLjB4qv
73vbEr/gN3NoCZW7Fr0PXRN8tdRL1b5POKbLPwr88CbR8B3jtd5i0tD/Ct3QUHkj
igV0t3F59jdPOQcx5E0CAwEAAaNCMEAwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB
/wQFMAMBAf8wHQYDVR0OBBYEFClW6q3+rnL6xyLQRelG5PsbjfhTMA0GCSqGSIb3
DQEBCwUAA4IBAQBWHmnjX/HiKvyVpOJU7c4vdIglWVKYJzYBWWJkDaots2UbPmOJ
bxIpbHyQoGsnViPDsSSPlIeogY+SspCOcweUXjqKYciKuZHTPDiKbXjY5rzTfrp0
DE6RU0gJMMAA9IwXkxZb50qi9XD9OOBc3uQhCEAddzr2DqOgoZdCIvNJnD8fOW+H
mpJEhO3Y0rPGXUy+bP8vUkhvfKXTQCzot23Kh+fG6oUNkBG4Qz0i0LWRKkXWu14n
SkCjsQng0U1BoS1UZn6YzpHrYnnVL7NyQpMbBxcShKwyOzeTcHyyUBUT16OKa87G
XAGBJaThvgbGE456DwEpRksSvCey91bC7qBD
-----END CERTIFICATE-----
```
**scheduler 证书用法与 controller-manager 一致**
`此处分析了apiserver、kube-controller-manager 两个服务在使用证书上的区别，但还有一个关键证书没有分析到，那就是 kubelet`
```
查看kublet证书文件
[root@dev01 pki]# cd /var/lib/kubelet/pki
[root@dev01 pki]# tree
.
├── kubelet-client-current.pem 
├── kubelet.crt
├── kubelet.key

[root@dev01 pki]# ls /etc/kubernetes
kubelet.conf
```
**我们可以看到 kubelet 有四个文件 kubelet-client-current.pem、kubelet.crt、kubelet.key、kubelet.conf**
`首先我们需要肯定 kubelet 也是通过 kubelet.conf 与 apiserver 通信`
`区别是 kubelet 使用到了 controller 给他生成的 kubelet-client-current.pem`

**查看 kubelet.conf 文件**
```
[root@dev01 kubernetes]# cat kubelet.conf
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM1ekNDQWMrZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeU1EVXlOREF6TVRZek1Wb1hEVE15TURVeU1UQXpNVFl6TVZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBUER4CnFjQitOZlJjY3hXY1l0N2ltREVLTVN2UnU5RUQ3K3Zha1pMeURuL0JKQmNQSXNRVDdnVHhhZkwxQVZlMWY4dU4KY09wRE9uNmgwUWxHVzI4WXpPYjBUSnJ6TEs1Q2ZHOWpvS0p0emxXdkJrU2tiNTV5enYySWlWTXJMNW9pbmJ1UAo0dGVKeCtmNWFKQlRiRFBDN2hjSTRaTDNPUXAwc0Q5OXZTbFNOeVcxTi9tZDFmRjJIY2dmcEdaZjNhUEFnSW1yCnp6TFJka2xDNVdHcGxRZFNzWlhheUxoYUIzOXFPdDZVQ2NCSVI4R0lVKzhQZDdkNGpOUEpVUW5oUFNMakI0cXYKNzN2YkVyL2dOM05vQ1pXN0ZyMFBYUk44dGRSTDFiNVBPS2JMUHdyODhDYlI4QjNqdGQ1aTB0RC9DdDNRVUhragppZ1YwdDNGNTlqZFBPUWN4NUUwQ0F3RUFBYU5DTUVBd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZDbFc2cTMrcm5MNnh5TFFSZWxHNVBzYmpmaFRNQTBHQ1NxR1NJYjMKRFFFQkN3VUFBNElCQVFCV0htbmpYL0hpS3Z5VnBPSlU3YzR2ZElnbFdWS1lKellCV1dKa0Rhb3RzMlViUG1PSgpieElwYkh5UW9Hc25WaVBEc1NTUGxJZW9nWStTc3BDT2N3ZVVYanFLWWNpS3VaSFRQRGlLYlhqWTVyelRmcnAwCkRFNlJVMGdKTU1BQTlJd1hreFpiNTBxaTlYRDlPT0JjM3VRaENFQWRkenIyRHFPZ29aZENJdk5KbkQ4Zk9XK0gKbXBKRWhPM1kwclBHWFV5K2JQOHZVa2h2ZktYVFFDem90MjNLaCtmRzZvVU5rQkc0UXowaTBMV1JLa1hXdTE0bgpTa0Nqc1FuZzBVMUJvUzFVWm42WXpwSHJZbm5WTDdOeVFwTWJCeGNTaEt3eU96ZVRjSHl5VUJVVDE2T0thODdHClhBR0JKYVRodmdiR0U0NTZEd0VwUmtzU3ZDZXk5MWJDN3FCRAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
    server: https://127.0.0.1:6443
  name: cluster.local
contexts:
- context:
    cluster: cluster.local
    user: system:node:dev01
  name: system:node:dev01@cluster.local
current-context: system:node:dev01@cluster.local
kind: Config
preferences: {}
users:
- name: system:node:dev01
  user:
    client-certificate: /var/lib/kubelet/pki/kubelet-client-current.pem
    client-key: /var/lib/kubelet/pki/kubelet-client-current.pem
```
`可以看出 kubelet.conf 与 kube-controller-manager.conf 不同的地方在于 users 不同，kube-controller-manager users 填充的是密钥而 kubelet 填充的是文件路径`

`kubeadm init 时，kubeadm 会创建临时 token，来创建 bootstrap-kubelet.conf，然后使用该凭证启动 kubelet，kubelet 会自动向 api-server 提交自签的 kubelet.crt、kubelet.key，controller-manager 会自动审批并下发给 kubelet-client-current.pem，并生成 kubelet.conf，删除之前bootstrap-kubelet.conf 重新启动 kubelet。当然 kubadm join 时也是类似原理。`
### 轮换证书
**[kubernetes轮换证书可以分为三个部分](https://kubernetes.io/zh-cn/docs/tasks/tls/)**

* 管理集群中的 TLS 认证
* 手动轮换 CA 证书
* 为 kubelet 配置证书轮换

**为 kubelet 配置证书轮换**

`需要1.8.0之后版本使用该特性`
`Kubernetes 包含特性 kubelet 证书轮换， 在当前证书即将过期时， 将自动生成新的秘钥，并从 Kubernetes API 申请新的证书。 一旦新的证书可用，它将被用于与 Kubernetes API 间的连接认证`

`启用客户端证书轮换
kubelet 进程接收 --rotate-certificates 参数，该参数决定 kubelet 在当前使用的证书即将到期时，是否会自动申请新的证书。
kube-controller-manager 进程接收 --cluster-signing-duration 参数 （在 1.19 版本之前为 --experimental-cluster-signing-duration），用来控制签发证书的有效期限，默认 1 年。`

**查看kubelet证书过期时间，然后设置时间，使其快要过期，查看系统是否自动续签**
```
## 配置 kubelet-config.yaml 自动续签特性
[root@dev02]# vi /etc/kubernetes/kubelet-config.yaml
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
...
rotateCertificates: true
...

## 配置 kube-controller-manager.yaml 控制时间有效期限
[root@dev02 pki]# vi /etc/kubernetes/manifests/kube-controller-manager.yaml
apiVersion: v1
kind: Pod
metadata:
...
spec:
  containers:
  - command:
    - kube-controller-manager
    - --cluster-signing-duration=87600h
   
```
**实验验证**
```
[root@dev02 pki]# openssl x509 -noout -in kubelet-client-current.pem  -dates
notBefore=Jul  8 07:32:24 2022 GMT
notAfter=Jul  8 07:32:24 2023 GMT

[root@dev02 pki]# date
2022年 07月 08日 星期五 16:29:54 CST
[root@dev02 pki]# date -s 2023-07-07
2023年 07月 07日 星期五 00:00:00 CST

-rw-------. 1 root root 2806 5月  24 2022 kubelet-client-2022-05-24-11-16-34.pem
-rw-------. 1 root root 2806 8月  10 2023 kubelet-client-2023-07-07-00-19-20.pem
lrwxrwxrwx. 1 root root   59 8月  10 2023 kubelet-client-current.pem -> /var/lib/kubelet/pki/kubelet-client-2023-07-07-00-19-20.pem
-rw-r--r--. 1 root root 2229 8月  10 2023 kubelet.crt
-rw-------. 1 root root 1679 8月  10 2023 kubelet.key
```
### 轮转失败
**如果kubelet自动续签轮转失败，可以手动轮转**
```
## 备份原数据
[root@dev01]# cd /etc/kubernetes/
[root@dev01 kubernetes]# cp kubelet.conf kubelet.conf.bak
[root@dev01 kubernetes]# cd /var/lib/kubelet/pki/
[root@dev01 pki]# mkdir bak
[root@dev01 pki]# mv * bak

## 生成新的 kubelet.conf
[root@dev02 pki]# kubeadm alpha kubeconfig user --config kubeadm-config.yaml --client-name=system:node:dev01 --org=system:nodes > kubelet.conf
## 重启
[root@dev02 pki]# systemctl restart kubelet
```
### 证书文件更新
**kubernetes 证书文件更新**
`检查证书文件到期时间，发现证书文件也是一年一签，但是证书文件没有自动续签功能，所以需要手动续签，kubeadm 已经提供了此功能`
```
[root@dev02 pki]# kubeadm certs check-expiration
CERTIFICATE                         EXPIRES                  RESIDUAL TIME   CERTIFICATE AUTHORITY   EXTERNALLY MANAGED
admin.conf                          Aug 08, 2024 16:03 UTC   <invalid>                               no
apiserver                           Aug 08, 2024 16:03 UTC   <invalid>       ca                      no
!MISSING! apiserver-etcd-client
apiserver-kubelet-client            Aug 08, 2024 16:03 UTC   <invalid>       ca                      no
controller-manager.conf             Aug 08, 2024 16:03 UTC   <invalid>                               no
!MISSING! etcd-healthcheck-client
!MISSING! etcd-peer
!MISSING! etcd-server
front-proxy-client                  Aug 08, 2024 16:03 UTC   <invalid>       front-proxy-ca          no
scheduler.conf                      Aug 08, 2024 16:03 UTC   <invalid>                               no

CERTIFICATE AUTHORITY   EXPIRES                  RESIDUAL TIME   EXTERNALLY MANAGED
ca                      May 21, 2032 03:16 UTC   7y              no
!MISSING! etcd-ca
front-proxy-ca          May 21, 2032 03:16 UTC   7y              no
```
**续签证书**
```
[root@dev02 pki]# kubeadm certs renew all
[renew] Reading configuration from the cluster...
[renew] FYI: You can look at this config file with 'kubectl -n kube-system get cm kubeadm-config -o yaml'
W0807 00:02:39.283691 1069553 utils.go:69] The recommended value for "clusterDNS" in "KubeletConfiguration" is: [10.232.0.10]; the provided value is: [169.254.25.10]

certificate embedded in the kubeconfig file for the admin to use and for kubeadm itself renewed
certificate for serving the Kubernetes API renewed
certificate for the API server to connect to kubelet renewed
certificate embedded in the kubeconfig file for the controller manager to use renewed
certificate for the front proxy client renewed
certificate embedded in the kubeconfig file for the scheduler manager to use renewed

Done renewing certificates. You must restart the kube-apiserver, kube-controller-manager, kube-scheduler and etcd, so that they can use the new certificates.
```
**查看证书过期时间**
```
[root@dev02 pki]# kubeadm certs check-expiration
[check-expiration] Reading configuration from the cluster...
[check-expiration] FYI: You can look at this config file with 'kubectl -n kube-system get cm kubeadm-config -o yaml'
W0807 00:02:52.800009 1069679 utils.go:69] The recommended value for "clusterDNS" in "KubeletConfiguration" is: [10.232.0.10]; the provided value is: [169.254.25.10]

CERTIFICATE                EXPIRES                  RESIDUAL TIME   CERTIFICATE AUTHORITY   EXTERNALLY MANAGED
admin.conf                 Aug 06, 2025 16:02 UTC   364d                                    no
apiserver                  Aug 06, 2025 16:02 UTC   364d            ca                      no
apiserver-kubelet-client   Aug 06, 2025 16:02 UTC   364d            ca                      no
controller-manager.conf    Aug 06, 2025 16:02 UTC   364d                                    no
front-proxy-client         Aug 06, 2025 16:02 UTC   364d            front-proxy-ca          no
scheduler.conf             Aug 06, 2025 16:02 UTC   364d                                    no

CERTIFICATE AUTHORITY   EXPIRES                  RESIDUAL TIME   EXTERNALLY MANAGED
ca                      May 21, 2032 03:16 UTC   7y              no
front-proxy-ca          May 21, 2032 03:16 UTC   7y              no
```

`手动更新证书文件不包括 kubelet，所以 kubelet 可以等待让其自动更新，或者手动更新`
`自动更新只能更新 kubelet，所以 apiserver 之类证书快过期时还需要人工手动操作`
`手动更新证书之后，需要移除静态 pod 让其重新创建，因为静态 pod 里关于证书部分还不能动态发现`