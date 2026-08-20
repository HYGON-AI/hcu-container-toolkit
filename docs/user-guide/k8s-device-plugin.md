# 在 HCU Device Plugin 中感知 HCU 设备

HCU Device Plugin 可正常完成 Kubernetes 集群中 HCU/vHCU 设备资源分配，但其运行需要特权模式支持。借助 HCU Container Toolkit，可解决此限制，简化部署。

> [!TIP]
> 关于 HCU Device Plugin 的详细介绍可查阅[开发者社区文档](https://developer.sourcefind.cn/document/87ee5c5b-c10d-11f0-b077-0242ac150003?id=8df80ff9-c10e-11f0-b077-0242ac150003)

## 配置写入

根据 Kubernetes 集群使用的容器运行时，执行对应命令调整其配置。

```bash
# 若集群所用容器运行时为Docker，则只需重启Docker服务
systemctl restart docker
# 若集群所有容器运行时为Containerd，执行如下命令
hcu-ctk runtime configure --runtime=containerd --set-as-default
systemctl restart containerd
```

## 工作模式

### 标准模式

该模式的编排文件内容如下。

<details><summary>k8s-hcu-plugin.yaml</summary>

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
      - image: <HCU_REPOSITORY>/hcu-device-plugin:v2.4.0
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

</details>

执行`kubectl apply -f k8s-hcu-plugin.yaml`完成部署并在 Pod 正常启动后，查看节点描述信息以确认生效。

<details><summary>节点描述信息</summary>

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

</details>

### MIG 模式
> [!TIP]
> MIG 设备创建、销毁等操作可参考[开发者社区文档](https://developer.sourcefind.cn/document/9169ef18-c10d-11f0-b077-0242ac150003?id=4a82aeed-e242-11f0-b9e4-0242ac150003)

将使用 MIG 设备的节点添加上标签，并启用 MIG 模式，创建实例。

```bash
kubectl label no <MIG_NODE> hcu-mode=mig
```

该模式的编排文件内容如下。

<details><summary>k8s-hcu-plugin-mig.yaml</summary>

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
        - image: <HCU_REPOSITORY>/hcu-device-plugin:v2.4.0
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

</details>

执行`kubectl apply -f k8s-hcu-plugin-mig.yaml`完成部署并在 Pod 正常启动后，查看节点描述信息以确认生效。

<details><summary>节点描述信息</summary>

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

</details>

### vHCU 动态切分模式

将启用动态切分模式的节点添加上标签。

```bash
kubectl label no <MIG_NODE> hcu=on
```

该模式的编排文件内容如下。

<details><summary>k8s-hcu-plugin-hami.yaml</summary>

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
      - image: <HCU_REPOSITORY>/hcu-device-plugin:v2.4.0
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

</details>

执行`kubectl apply -f k8s-hcu-plugin-hami.yaml`完成部署并在 Pod 正常启动后，查看节点描述信息以确认生效。
另外，参考[开发者社区文档](https://developer.sourcefind.cn/document/9169ef18-c10d-11f0-b077-0242ac150003?id=b782480c-e235-11f0-b9e4-0242ac150003)完成 vHCU-Scheduler 安装部署。

<details><summary>节点描述信息</summary>

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
                    hygon.cn/node-hcu-register:
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

</details>

