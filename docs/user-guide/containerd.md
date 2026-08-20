# 在 Containerd 容器中挂载 HCU 设备

## 配置写入

执行如下命令，相关配置会写入到`/etc/containerd/config.toml`中。

```bash
hcu-ctk runtime configure --runtime=containerd --set-as-default
systemctl restart containerd
```

## 运行时集成

> [!TIP]
> Linux 普通用户需使用 sudo 提升权限运行容器

### 使用环境变量挂载整卡

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

### 使用 CDI 方式挂载整卡
> [!TIP]
> 需 Containerd 1.7+版本支持

此方式需要生成 CDI spec 文件，并确保配置中启用 CDI。

```bash
hcu-ctk cdi generate --output=/etc/cdi/hcu.yaml
# Containerd 2.0+版本已默认启用CDI
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

### 使用环境变量挂载虚拟卡
> [!TIP]
> 虚拟卡创建、销毁等操作可参考[开发者社区文档](https://developer.sourcefind.cn/document/9169ef18-c10d-11f0-b077-0242ac150003?id=b782480c-e235-11f0-b9e4-0242ac150003)

通过`-e`设置`VHCU_VISIBLE_DEVICES`环境变量，变量值可为节点单个 vHCU（例如 0）、多个 vHCU（例如 0,4）。

- 挂载单个 vHCU
```bash
nerdctl run -it --runtime hcu-container-runtime -e VHCU_VISIBLE_DEVICES=0 <IMAGE>
```

- 挂载多个 vHCU
```bash
# 不可同时挂载同一物理卡切分出来的vHCU
nerdctl run -it --runtime hcu-container-runtime -e VHCU_VISIBLE_DEVICES=0,4 <IMAGE>
```
