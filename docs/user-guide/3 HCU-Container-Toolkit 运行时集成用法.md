# 3 HCU-Container-Toolkit 运行时集成用法

## 3.1 在 Docker 容器中挂载 HCU 设备

### 3.1.1 配置写入

HCU Container Toolkit 安装完成后，已经自动执行了针对 Docker 配置调整的命令，因此，仅需重启 Docker 服务即可。

```bash
systemctl daemon-reload
systemctl restart docker
```

### 3.1.2 运行时集成

若需启用 RDMA 网络，可通过`-e`设置`HCU_MOFED=enabled`

#### 3.1.2.1 使用环境变量挂载整卡

通过`-e`设置`HCU_VISIBLE_DEVICES`环境变量，变量值可为节点所有 HCU（all）、单个 HCU（例如 0）、多个 HCU（例如 0,1）。若设置`ROCM_VERSION`环境变量，则默认挂载节点所有 HCU。

- 挂载所有 HCU

```bash
docker run -it -e HCU_VISIBLE_DEVICES=all <IMAGE>
```

- 挂载单个 HCU

```bash
docker run -it -e HCU_VISIBLE_DEVICES=0 <IMAGE>
```

- 挂载多个 HCU

```bash
docker run -it -e HCU_VISIBLE_DEVICES=0,1 <IMAGE>
```

#### 3.1.2.2 使用 CDI 方式挂载整卡

> [!TIP]  
> 需 Docker 25.0+版本支持

此方式需要生成 CDI spec 文件，并配置启用 CDI。

```bash
hcu-ctk cdi generate --output=/etc/cdi/hcu.yaml
hcu-ctk runtime configure --runtime=docker --set-as-default --cdi.enabled
systemctl restart docker
```

可通过如下命令查看已生成的 CDI 设备。

```bash
hcu-ctk cdi list
```

通过`--device`指定 HCU 设备，其值可为节点所有 HCU（`hygon.com/hcu=all`）、单个 HCU（例如`hygon.com/hcu=0`或`hygon.com/hcu=hcu-T6T8290017030301`）。

- 挂载所有 HCU

```bash
docker run -it --device hygon.com/hcu=all <IMAGE>
```

- 挂载单个 HCU

```bash
docker run -it --device hygon.com/hcu=0 <IMAGE>
```

- 挂载多个 HCU

```bash
docker run -it --device hygon.com/hcu=0 --device hygon.com/hcu=1 <IMAGE>
```

#### 3.1.2.3 使用--gpus 挂载整卡

> [!TIP]  
> 需 Docker 19.0+版本支持

`--gpus`参数的值可以为节点全部 HCU（all）、单个 HCU（写法为`device=0`）、多个 HCU（写法为`'"device=0,2"'`）、指定 HCU 的个数（例如 2）。

- 挂载所有 HCU

```bash
docker run -it --gpus all <IMAGE>
```

- 挂载单个 HCU

```bash
docker run -it --gpus 1 <IMAGE>
docker run -it --gpus device=0 <IMAGE>
```

- 挂载多个 HCU

```bash
docker run -it --gpus 2 <IMAGE>
docker run -it --gpus '"device=0,1"' <IMAGE>
```

#### 3.1.2.4 使用环境变量挂载虚拟卡

> [!TIP]  
> 虚拟卡创建、销毁等操作可参考[开发者社区文档](https://developer.sourcefind.cn/document/9169ef18-c10d-11f0-b077-0242ac150003?id=b782480c-e235-11f0-b9e4-0242ac150003)

通过`-e`设置`VHCU_VISIBLE_DEVICES`环境变量，变量值可为节点单个 vHCU（例如 0）、多个 vHCU（例如 0,4）。

- 挂载单个 vHCU

```bash
docker run -it -e VHCU_VISIBLE_DEVICES=0 <IMAGE>
```

- 挂载多个 vHCU

```bash
# 不可同时挂载同一物理卡切分出来的vHCU
docker run -it -e VHCU_VISIBLE_DEVICES=0,4 <IMAGE>
```

## 3.2 在 Containerd 容器中挂载 HCU 设备

### 3.2.1 配置写入

执行如下命令，相关配置会写入到`/etc/containerd/config.toml`中。

```bash
hcu-ctk runtime configure --runtime=containerd --set-as-default
systemctl daemon-reload
systemctl restart containerd
```

### 3.2.2 运行时集成

> [!TIP]  
> Linux 普通用户需使用 sudo 提升权限运行容器

若需启用 RDMA 网络，可通过-e 设置 HCU_MOFED=enabled

#### 3.2.2.1 使用环境变量挂载整卡

通过`-e`设置`HCU_VISIBLE_DEVICES`环境变量，变量值可为节点所有 HCU（all）、单个 HCU（例如 0）、多个 HCU（例如 0,1）。若设置`ROCM_VERSION`环境变量，则默认挂载节点所有 HCU。

- 挂载所有 HCU

```bash
nerdctl run -it --runtime hcu-container-runtime -e HCU_VISIBLE_DEVICES=all <IMAGE>
```

- 挂载单个 HCU

```bash
nerdctl run -it --runtime hcu-container-runtime -e HCU_VISIBLE_DEVICES=0 <IMAGE>
```

- 挂载多个 HCU

```bash
nerdctl run -it --runtime hcu-container-runtime -e HCU_VISIBLE_DEVICES=0,1 <IMAGE>
```

#### 3.2.2.2 使用 CDI 方式挂载整卡

> [!TIP]  
> 需 Containerd 1.7+版本支持

此方式需要生成 CDI spec 文件，并确保配置中启用 CDI。

```bash
hcu-ctk cdi generate --output=/etc/cdi/hcu.yaml
# Containerd 2.0+ 版本已默认启用 CDI
hcu-ctk runtime configure --runtime=containerd --set-as-default --cdi.enabled
```

可通过如下命令查看已生成的 CDI 设备。

```bash
hcu-ctk cdi list
```

通过`--device`指定 HCU 设备，其值可为节点所有 HCU（`hygon.com/hcu=all`）、单个 HCU（例如`hygon.com/hcu=0`或`hygon.com/hcu=hcu-T6T8290017030301`）

- 挂载所有 HCU

```bash
nerdctl run -it --runtime hcu-container-runtime --device hygon.com/hcu=all <IMAGE>
```

- 挂载单个 HCU

```bash
nerdctl run -it --runtime hcu-container-runtime --device hygon.com/hcu=0 <IMAGE>
```

- 挂载多个 HCU

```bash
nerdctl run -it --runtime hcu-container-runtime --device hygon.com/hcu=0 --device hygon.com/hcu=1 <IMAGE>
```

#### 3.2.2.3 使用环境变量挂载虚拟卡

> [!TIP]  
> 虚拟卡创建、销毁等操作可参考[开发者社区文档](https://developer.sourcefind.cn/document/9169ef18-c10d-11f0-b077-0242ac150003?id=b782480c-e235-11f0-b9e4-0242ac150003)

通过`-e`设置`VHCU_VISIBLE_DEVICES`环境变量，变量值可为节点单个 vHCU（例如 0）、多个 vHCU（例如 0,4）。

- 挂载单个 vHCU

```bash
nerdctl run -it --runtime hcu-container-runtime -e VHCU_VISIBLE_DEVICES=0 <IMAGE>
```

- 挂载多个 vHCU

```bash
# 不可同时挂载同一物理卡切分出来的 vHCU
nerdctl run -it --runtime hcu-container-runtime -e VHCU_VISIBLE_DEVICES=0,4 <IMAGE>
```

## 3.3 在 Podman 容器中挂载 HCU 设备

### 3.3.1 配置写入

执行如下命令，相关配置会写入到`/etc/containers/containers.conf`中。

```bash
hcu-ctk runtime configure --runtime=podman --set-as-default
```

### 3.3.2 运行时集成

> [!TIP]  
> Linux 普通用户需使用 sudo 提升权限运行容器

若需启用 RDMA 网络，可通过-e 设置 HCU_MOFED=enabled

#### 3.3.2.1 使用环境变量挂载整卡

通过`-e`设置`HCU_VISIBLE_DEVICES`环境变量，变量值可为节点所有 HCU（all）、单个 HCU（例如 0）、多个 HCU（例如 0,1）。若设置`ROCM_VERSION`环境变量，则默认挂载节点所有 HCU。

- 挂载所有 HCU

```bash
podman run -it -e HCU_VISIBLE_DEVICES=all <IMAGE>
```

- 挂载单个 HCU

```bash
podman run -it -e HCU_VISIBLE_DEVICES=0 <IMAGE>
```

- 挂载多个 HCU

```bash
podman run -it -e HCU_VISIBLE_DEVICES=0,1 <IMAGE>
```

#### 3.3.2.2 使用 CDI 方式挂载整卡

> [!TIP]  
> 需 Podman 3.2+版本支持

此方式需要生成 CDI spec 文件。

```bash
hcu-ctk cdi generate --output=/etc/cdi/hcu.yaml
```

可通过如下命令查看已生成的 CDI 设备。

```bash
hcu-ctk cdi list
```

通过`--device`指定 HCU 设备，其值可为节点所有 HCU（`hygon.com/hcu=all`）、单个 HCU（例如`hygon.com/hcu=0`或`hygon.com/hcu=hcu-T6T8290017030301`）

- 挂载所有 HCU

```bash
podman run -it --device hygon.com/hcu=all <IMAGE>
```

- 挂载单个 HCU

```bash
podman run -it --device hygon.com/hcu=0 <IMAGE>
```

- 挂载多个 HCU

```bash
podman run -it --device hygon.com/hcu=0 --device hygon.com/hcu=1 <IMAGE>
```

#### 3.3.2.3 使用环境变量挂载虚拟卡

> [!TIP]  
> 虚拟卡创建、销毁等操作可参考[开发者社区文档](https://developer.sourcefind.cn/document/9169ef18-c10d-11f0-b077-0242ac150003?id=b782480c-e235-11f0-b9e4-0242ac150003)

通过`-e`设置`VHCU_VISIBLE_DEVICES`环境变量，变量值可为节点单个 vHCU（例如 0）、多个 vHCU（例如 0,4）。

- 挂载单个 vHCU

```bash
podman run -it -e VHCU_VISIBLE_DEVICES=0 <IMAGE>
```

- 挂载多个 vHCU

```bash
# 不可同时挂载同一物理卡切分出来的 vHCU
podman run -it -e VHCU_VISIBLE_DEVICES=0,4 <IMAGE>
```

## 3.4 在 HCU-Device-Plugin 中感知 HCU 设备

HCU Device Plugin 可正常完成 Kubernetes 集群中 HCU/vHCU 设备资源分配，但其运行需要特权模式支持。借助 HCU Container Toolkit，可解决此限制，简化部署。

> [!TIP]  
> 关于 HCU Device Plugin 的详细介绍可查阅[开发者社区文档](https://developer.sourcefind.cn/document/87ee5c5b-c10d-11f0-b077-0242ac150003?id=8df80ff9-c10e-11f0-b077-0242ac150003)

### 3.4.1 配置写入

根据 Kubernetes 集群使用的容器运行时，执行对应命令调整其配置。

```bash
# 若集群所用容器运行时为 Docker，则只需重启 Docker 服务
systemctl daemon-reload
systemctl restart docker
# 若集群所用容器运行时为 Containerd，执行如下命令
hcu-ctk runtime configure --runtime=containerd --set-as-default
systemctl daemon-reload
systemctl restart containerd
```

### 3.4.2 工作模式

#### 3.4.2.1 标准模式

该模式的编排文件内容如下。

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: hcu-device-plugin
  namespace: kube-system
spec:
  selector:
    matchLabels:
      name: hcu-dp-ds
  template:
    metadata:
      labels:
        name: hcu-dp-ds
    spec:
      tolerations:
      - key: CriticalAddonsOnly
        operator: Exists
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: hygon.com/hcu
                operator: In
                values:
                - "true"
              - key: hcu-mode
                operator: NotIn
                values:
                - mig
              - key: hcu
                operator: NotIn
                values:
                - "on"
      containers:
      - image: harbor.sourcefind.cn:5443/hcu/admin/base/hcu-device-plugin:v2.4.0
        name: hcu-dp-cntr
        env:
          - name: "HCU_VISIBLE_DEVICES"
            value: "all"
          - name: "PULSE"
            value: "30"
          - name: "RESOURCE_REGISTER_STRATEGY"
            value: "mixed"
          - name: "LOG_THRESHOLD"
            value: "INFO"
          - name: "LOG_VERBOSE"
            value: "2"
          - name: "LOG_OUTPUT"
            value: "true"
          - name: "POLICY"
            value: "0"
          - name: "NODE_NAME"
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
        volumeMounts:
          - name: dp
            mountPath: /var/lib/kubelet/device-plugins
      volumes:
        - name: dp
          hostPath:
            path: /var/lib/kubelet/device-plugins
```

执行`kubectl apply -f k8s-hcu-plugin.yaml`完成部署并在 Pod 正常启动后，查看节点描述信息以确认生效。

```text
[root@worker1 ~]# kubectl describe no worker1 
Name:               worker1
Roles:              worker
Labels:             beta.kubernetes.io/arch=amd64
                    beta.kubernetes.io/os=linux
                    hygon.com/hcu=true
                    hygon.com/hcu.cu-count=120
                    hygon.com/hcu.name=K100_AI
                    hygon.com/hcu.vram=64G
                    kubernetes.io/arch=amd64
                    kubernetes.io/hostname=worker1
                    kubernetes.io/os=linux
                    node-role.kubernetes.io/worker=
...
Capacity:
  cpu:                32
  ephemeral-storage:  226604556Ki
  hugepages-1Gi:      0
  hugepages-2Mi:      0
  hygon.com/hcu:      3
  memory:             263750076Ki
  pods:               110
Allocatable:
  cpu:                31600m
  ephemeral-storage:  226604556Ki
  hugepages-1Gi:      0
  hugepages-2Mi:      0
  hygon.com/hcu:      3
  memory:             256051785732
  pods:               110
...
```

#### 3.4.2.2 MIG 模式

> [!TIP]  
> MIG 设备创建、销毁等操作可参考[开发者社区文档](https://developer.sourcefind.cn/document/9169ef18-c10d-11f0-b077-0242ac150003?id=4a82aeed-e242-11f0-b9e4-0242ac150003)

将使用 MIG 设备的节点添加上标签，并启用 MIG 模式，创建实例。

```bash
kubectl label no <MIG_NODE> hcu-mode=mig
```

该模式的编排文件内容如下。

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: hcu-device-plugin-mig
  namespace: kube-system
spec:
  selector:
    matchLabels:
      name: hcu-dp-ds-mig
  template:
    metadata:
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ""
      labels:
        name: hcu-dp-ds-mig
    spec:
      nodeSelector:
        hcu-mode: mig
        hygon.com/hcu: "true"
      tolerations:
        - key: CriticalAddonsOnly
          operator: Exists
      containers:
        - image: harbor.sourcefind.cn:5443/hcu/admin/base/hcu-device-plugin:v2.4.0
          name: hcu-dp-cntr
          env:
            - name: "HCU_VISIBLE_DEVICES"
              value: "all"
            - name: "PULSE"
              value: "30"
            - name: "RESOURCE_REGISTER_STRATEGY"
              value: "mig"
            - name: "LOG_THRESHOLD"
              value: "INFO"
            - name: "LOG_VERBOSE"
              value: "2"
            - name: "LOG_OUTPUT"
              value: "true"
            - name: "NODE_NAME"
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          volumeMounts:
            - name: dp
              mountPath: /var/lib/kubelet/device-plugins
      volumes:
        - name: dp
          hostPath:
            path: /var/lib/kubelet/device-plugins
```

执行`kubectl apply -f k8s-hcu-plugin-mig.yaml`完成部署并在 Pod 正常启动后，查看节点描述信息以确认生效。

```text
[root@worker1 ~]# kubectl describe no worker1 
Name:               worker1
Roles:              worker
Labels:             beta.kubernetes.io/arch=amd64
                    beta.kubernetes.io/os=linux
                    hcu-mode=mig
                    hygon.com/hcu=true
                    hygon.com/hcu.cu-count=120
                    hygon.com/hcu.name=K100_AI
                    hygon.com/hcu.vram=64G
                    kubernetes.io/arch=amd64
                    kubernetes.io/hostname=worker1
                    kubernetes.io/os=linux
                    node-role.kubernetes.io/worker=
...
Capacity:
  cpu:                          32
  ephemeral-storage:            226604556Ki
  hugepages-1Gi:                0
  hugepages-2Mi:                0
  hygon.com/hcu-mig-2g-15gb:    1
  memory:                       263750076Ki
  pods:                         110
Allocatable:
  cpu:                          31600m
  ephemeral-storage:            226604556Ki
  hugepages-1Gi:                0
  hugepages-2Mi:                0
  hygon.com/hcu-mig-2g-15gb:    1
  memory:                       256051785732
  pods:                         110
...
```

#### 3.4.2.3 vHCU 动态切分模式

将启用动态切分模式的节点添加上标签。

```bash
kubectl label no <MIG_NODE> hcu=on
```

该模式的编排文件内容如下。

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: hcu-device-plugin-hami
  namespace: kube-system
spec:
  selector:
    matchLabels:
      name: hcu-dp-ds-hami
  template:
    metadata:
      labels:
        name: hcu-dp-ds-hami
    spec:
      tolerations:
      - key: CriticalAddonsOnly
        operator: Exists
      nodeSelector:
        hcu: "on"
      serviceAccountName: hcu-device-plugin
      containers:
      - image: harbor.sourcefind.cn:5443/hcu/admin/base/hcu-device-plugin:v2.4.0
        name: hcu-dp-cntr
        env:
          - name: "HCU_VISIBLE_DEVICES"
            value: "all"
          - name: "PULSE"
            value: "30"
          - name: "RESOURCE_REGISTER_STRATEGY"
            value: "hami"
          - name: "LOG_THRESHOLD"
            value: "INFO"
          - name: "LOG_OUTPUT"
            value: "true"
          - name: "LOG_VERBOSE"
            value: "2"
          - name: "POLICY"
            value: "0"
          - name: "RESOURCE_MULTIPLE"
            value: "false"
          - name: "NODE_NAME"
            valueFrom:
              fieldRef:
                fieldPath: spec.nodeName
        volumeMounts:
          - name: dp
            mountPath: /var/lib/kubelet/device-plugins
      volumes:
        - name: dp
          hostPath:
            path: /var/lib/kubelet/device-plugins

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: hcu-device-plugin
rules:
  - apiGroups:
      - ""
    resources:
      - nodes
    verbs:
      - get
      - update
      - list
      - patch
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - update
      - patch
      - get
      - list
      - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: hcu-device-plugin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: hcu-device-plugin
subjects:
  - kind: ServiceAccount
    name: hcu-device-plugin
    namespace: kube-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: hcu-device-plugin
  namespace: kube-system
```

执行`kubectl apply -f k8s-hcu-plugin-hami.yaml`完成部署并在 Pod 正常启动后，查看节点描述信息以确认生效。  
另外，参考[开发者社区文档](https://developer.sourcefind.cn/document/9169ef18-c10d-11f0-b077-0242ac150003?id=b782480c-e235-11f0-b9e4-0242ac150003)完成 vHCU-Scheduler 安装部署。

```text
[root@worker1 ~]# kubectl describe no worker1 
Name:               worker1
Roles:              worker
Labels:             beta.kubernetes.io/arch=amd64
                    beta.kubernetes.io/os=linux
                    hcu=on
                    hygon.com/hcu=true
                    hygon.com/hcu.cu-count=120
                    hygon.com/hcu.name=K100_AI
                    hygon.com/hcu.vram=64G
                    kubernetes.io/arch=amd64
                    kubernetes.io/hostname=worker1
                    kubernetes.io/os=linux
                    node-role.kubernetes.io/worker=
Annotations:        ...
                    hami.io/node-hcu-register:
                      HCU-T6T8290019020401,4,65520,100,HCU-K100_AI,0,true,0,hami:HCU-T6T8290019030301,4,65520,100,HCU-K100_AI,1,true,1,hami:HCU-T6T8290017030301...
                    hami.io/node-handshake-hcu: Requesting_2026-01-28 11:25:27
                    hygon.com/node-hcu-register:
                      HCU-T6T8290019020401,4,65520,100,HCU-K100_AI,0,true,0,hami:HCU-T6T8290019030301,4,65520,100,HCU-K100_AI,1,true,1,hami:HCU-T6T8290017030301...
                    kubeadm.alpha.kubernetes.io/cri-socket: unix:///run/containerd/containerd.sock
                    node.alpha.kubernetes.io/ttl: 0
                    volumes.kubernetes.io/controller-managed-attach-detach: true
...
Capacity:
  cpu:                          32
  ephemeral-storage:            226604556Ki
  hugepages-1Gi:                0
  hugepages-2Mi:                0
  hygon.com/hcunum:             12
  memory:                       263750076Ki
  pods:                         110
Allocatable:
  cpu:                          31600m
  ephemeral-storage:            226604556Ki
  hugepages-1Gi:                0
  hugepages-2Mi:                0
  hygon.com/hcunum:             12
  memory:                       256051785732
  pods:                         110
...
```
