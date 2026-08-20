# 在 Docker 容器中挂载 HCU 设备

## 配置写入

HCU Container Toolkit 安装完成后，已经自动执行了针对 Docker 配置调整的命令，因此，仅需重启 Docker 服务即可。

```bash
systemctl restart docker
```

## 运行时集成

### 使用环境变量挂载整卡

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

### 使用 CDI 方式挂载整卡
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

### 使用--gpus 挂载整卡
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

### 使用环境变量挂载虚拟卡
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
