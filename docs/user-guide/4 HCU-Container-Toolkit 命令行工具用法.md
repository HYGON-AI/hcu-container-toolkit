# 4 HCU-Container-Toolkit 命令行工具用法

hcu-ctk 是进行运行时配置及管理 HCU 容器的命令行工具，可根据需求执行对应子命令。其各子命令使用可移步相应的文档。

## 4.1 运行时配置调整

hcu-ctk runtime configure 子命令提供了多项参数，以方便对指定运行时配置调整。可用参数如下：

- `--runtime`：目标容器运行时，可指定`docker`、`containerd`、`podman`，默认值为`docker`
- `--config`：容器运行时配置文件路径，仅在用户自行调整过配置文件路径的情况下指定
- `--hcu-runtime-name`：写入目标容器运行时配置的 HCU Runtime 命令，默认值为`hcu`
- `--hcu-runtime-path`或`--runtime-path`：HCU Container Runtime 二进制路径，默认值为`hcu-container-runtime`
- `--hcu-set-as-default`或`--set-as-default`：指定 HCU Runtime 为目标容器运行时配置中的默认运行时，默认值为`false`
- `--cdi.enabled`或`--cdi.enable`：在目标容器运行时配置中启用 CDI，默认值为`false`
- `--remove`：在目标容器运行时配置中移除 HCU Runtime
- `--rollback`：回滚目标容器运行时配置
- `--dry-run`：将调整输出到控制台，不写入到目标容器运行时配置文件中

示例用法如下：

```bash
# 在Docker运行时配置中加入HCU Runtime，并将其设置为默认运行时，且启用CDI
$ hcu-ctk runtime configure --runtime=docker --set-as-default --cdi.enabled
$ cat /etc/docker/daemon.json
{
    "debug": false,
    "default-runtime": "hcu",
    "features": {
        "cdi": true
    },
    "runtimes": {
        "hcu": {
            "args": [],
            "path": "hcu-container-runtime"
        }
    }
}

# 通过 dry-run 在控制台中查看待写入的配置
$ hcu-ctk runtime configure --runtime=docker --set-as-default --cdi.enabled --dry-run
WARN[0000] Ignoring runtime-config-override flag for docker 
INFO[0000] Loading config from /etc/docker/daemon.json  
{
    "debug": false,
    "default-runtime": "hcu",
    "features": {
        "cdi": true
    },
    "runtimes": {
        "hcu": {
            "args": [],
            "path": "hcu-container-runtime"
        }
    }
}

# 移除 Docker 配置文件中 HCU Runtime 部分
$ hcu-ctk runtime configure --runtime=docker --remove
WARN[0000] Ignoring runtime-config-override flag for docker 
INFO[0000] Loading config from /etc/docker/daemon.json  
INFO[0000] Wrote updated config to /etc/docker/daemon.json 
INFO[0000] It is recommended that docker daemon be restarted. 
$ cat /etc/docker/daemon.json
{
    "debug": false,
    "default-runtime": "runc",
    "features": {
        "cdi": true
    }
}

# 回滚 Docker 配置文件
$ hcu-ctk runtime configure --runtime=docker --rollback
WARN[0000] Ignoring runtime-config-override flag for docker 
INFO[0000] Restored the backup file /etc/docker/daemon.json.hcu-ctk.bak 
$ cat /etc/docker/daemon.json 
{
  "debug": false
}
```

## 4.2 CDI 管理

hcu-ctk cdi 子命令可以实现 CDI 规范生成、校验、展示等功能，以便在容器中访问和使用 CDI 设备。

> [!TIP]  
> CDI 介绍可参考[此地址](https://github.com/cncf-tags/container-device-interface)

### 4.2.1 cdi list

hcu-ctk cdi list 子命令可通过已生成的 CDI 规范展示相应 CDI 设备。该子命令提供了`--spec-dir`参数指定 CDI 规范文件所在目录，参数默认值为`“/etc/cdi,/var/run/cdi”`。

示例用法如下：

```bash
# 从默认目录读取 CDI 规范文件
$ hcu-ctk cdi list
INFO[0000] Found 1 CDI devices                          
hygon.com/hcu=0
hygon.com/hcu=all
hygon.com/hcu=hcu-TPXS300002100601

# 指定文件目录读取
$ hcu-ctk cdi list --spec-dir /tmp
INFO[0000] Found 1 CDI devices                          
hygon.com/hcu=0
hygon.com/hcu=all
hygon.com/hcu=hcu-TPXS300002100601
```

### 4.2.2 cdi generate

hcu-ctk cdi generate 子命令负责生成 CDI 规范文件，其可用参数如下：

- `--output`：CDI 规范写入文件路径，若设置空值或忽略则将 CDI 规范输出到控制台
- `--format`：CDI 规范的格式，可配置为`json`或`yaml`，默认值为`yaml`
- `--device-name-strategy`：指定设备名称的生成策略，可选`index`、`uuid`、`type-index`，默认值为`index,uuid`
- `--hcu-cdi-hook-path`或`--hcu-ctk-path`：指定 hcu-cdi-hook 二进制的路径，一般情况下无需指定
- `--vendor`或`--cdi-vendor`：自定义设备商 ID，默认值为`hygon.com`
- `--class`或`--cdi-class`：自定义设备类型，默认值为`hcu`

示例用法如下：

```bash
# 生成 CDI 规范，并写入到 /etc/cdi/hcu.yaml 中
$ hcu-ctk cdi generate --output /etc/cdi/hcu.yaml
$ hcu-ctk cdi list
INFO[0000] Found 1 CDI devices                          
hygon.com/hcu=0
hygon.com/hcu=all
hygon.com/hcu=hcu-TPXS300002100601

# 指定设备名称生成策略为 type-index 和 index
$ hcu-ctk cdi generate --device-name-strategy "type-index,index" --output /etc/cdi/hcu.yaml
$ hcu-ctk cdi list
INFO[0000] Found 1 CDI devices                          
hygon.com/hcu=0
hygon.com/hcu=all
hygon.com/hcu=hcu0
```

### 4.2.3 cdi validate

hcu-ctk cdi validate 子命令可对已生成的 CDI 规范文件进行校验，该功能适用于 HCU 硬件环境有变动的场景。该子命令提供了`--path`参数指定 CDI 规范文件路径，参数默认值为`/etc/cdi/hcu.yaml`。

示例用法如下：

```bash
# 校验默认路径的 CDI 规范文件
$ hcu-ctk cdi validate
...
CDI spec is valid

# 指定 CDI 规范文件路径
$ hcu-ctk cdi validate --path /tmp/hcu.yaml
...
CDI spec is valid
```

### 4.2.4 cdi transform

hcu-ctk cdi transform root 子命令可对 CDI 规范中挂载目录进行转换，该功能适用于驱动目录等有调整的场景。其可用参数如下：

- `--from`：指定待转换目录
- `--to`：指定转换后的目录
- `--relative-to`：指定转换的目录为本地目录或容器内目录，可选`host`和`container`，默认值为`host`
- `--input`：指定待转换文件，也可指定为`-`，代表在控制台输入 CDI 规范。默认值为`-`
- `--output`：指定转换后写入的文件，若未指定或指定为空，则将 CDI 规范输出到控制台

示例用法如下：

```bash
# 将 /etc/cdi/hcu.yaml 中挂载信息的本地目录 /opt/hyhal 调整为 /usr/local/hyhal，输出到控制台
hcu-ctk cdi transform root --from /opt/hyhal --to /usr/local/hyhal --relative-to host --input /etc/cdi/hcu.yaml

# 将 /etc/cdi/hcu.yaml 中挂载信息的本地目录 /opt/hyhal 调整为 /usr/local/hyhal，并写回到原文件
hcu-ctk cdi transform root --from /opt/hyhal --to /usr/local/hyhal --relative-to host --input /etc/cdi/hcu.yaml --output /etc/cdi/hcu.yaml
```

## 4.3 HCU-Container-Toolkit 配置管理

hcu-ctk config 子命令可对 HCU Container Toolkit 配置进行展示、生成、调整等操作，其可用参数如下：

- `--set`：设定待调整的配置项，若仅设定配置项的键，则其等同于`key=true`；若配置项的值为列表且有多个元素，则各元素之间使用`:`分隔
- `--config-file`或`--config`或`-c`：指定配置来源，默认值为`/etc/hcu-container-runtime/config.toml`
- `--in-place`或`-i`：指定是否将调整的配置项写回到来源文件，默认值为`false`；该参数和`--output`不可同时使用
- `--output`或`-o`：指定调整后配置写入的文件；若该参数未指定且同时未指定`--in-place`，调整后配置输出到控制台；该参数和`--in-place`不可同时使用

示例用法如下：

```bash
# 展示默认配置文件中的配置
hcu-ctk config

# 调整配置项，并将结果输出到控制台查看 
hcu-ctk config --set hcu-container-runtime.runtimes=runc:crun --set hcu-container-runtime.log-level=debug

# 调整配置项，并将结果写回到默认配置文件
hcu-ctk config --set hcu-container-runtime.debug="/var/log/hcu-container-runtime.log" -i

# 从指定文件获取配置并进行调整，并将结果写入到 /tmp/config.toml 中
hcu-ctk config -c /etc/hcu-container-runtime/config.toml --set hcu-container-runtime.log-level=info -o /tmp/config.toml
```

另外，hcu-ctk config default 子命令可展示、生成默认的 HCU Container Toolkit 配置。该子命令提供了`--output`（或`-o`）来指定默认配置的写入文件路径，若未指定，则将默认配置输出到控制台。

示例用法如下：

```bash
# 展示默认配置
hcu-ctk config default

# 将默认配置写入到指定文件中
hcu-ctk config default -o /etc/hcu-container-runtime/config.toml 
```

## 4.4 HCU 可访问性配置

HCU Tracker 可自行维护 HCU 状态及容器对 HCU 的可访问性配置。默认情况下 HCU Tracker 是 disable 状态，用户可自行配置启用。

> [!TIP]  
> 目前支持通过`--gpus`和`-e HCU_VISIBLE_DEVICES`方式启动的容器

### 4.4.1 启用/禁用 HCU Tracker

可通过 hcu-ctk 对 HCU Tracker 的启用或禁用进行调整。

```bash
# 功能启用
hcu-ctk hcu-tracker enable
# 功能禁用
hcu-ctk hcu-tracker disable
```

### 4.4.2 查看 HCU Tracker Status

在 HCU Tracker 功能启用后，可通过 status 查询 HCU 可访问性及对应容器的情况。

```bash
hcu-ctk hcu-tracker status
------------------------------------------------------------------------------------------------------------------------
HCU Id    UUID                     Accessibility       Container Ids                                                    
------------------------------------------------------------------------------------------------------------------------
0         0x7100F95F9B3230E1       Shared              b3277bc4c47b550d91ab97f40aa66e881a32e031a988d1443997ab3aea3fafa2 
1         0x7100F95F9B102101       Shared              None 
```

### 4.4.3 调整容器对 HCU 的访问权限

HCU Tracker 中 HCU 的可访问性有两种：

- `shared` 表示 HCU 可以同时被多个容器访问。默认情况下，所有 HCU 都被授予 shared 访问权限，以符合 Docker 的默认行为。
- `exclusive` 表示 HCU 在任何时刻最多只能被一个容器访问。

可通过如下命令调整权限。

```bash
# 调整多个 HCU 为 exclusive 权限
# 若 HCU 此时已被多个容器使用，则配置失败
hcu-ctk hcu-tracker 0-3 exclusive

# 调整单个 HCU 为 exclusive 权限
hcu-ctk hcu-tracker 4 exclusive

# 调整所有 HCU 为 shared 权限
hcu-ctk hcu-tracker all shared
```

### 4.4.4 重置 HCU Tracker

在 HCU 进行虚拟化操作后，应重置 HCU Tracker。

```bash
hcu-ctk hcu-tracker reset
```

## 4.5 Docker 容器中文件读写权限调整

在 Docker 默认的 rootful 模式下，Linux 普通用户启动的容器在使用`-v`挂载目录时可以赋予目录写入权限，即使该目录是其他 Linux 用户目录也可在容器内对目录下文件进行删增修改等操作。为规避这一问题带来的安全风险，可执行`hcu-ctk rootless`命令进行限制。

> [!TIP]  
> 该子命令仅用于 Docker 容器运行时

示例用法如下：

```bash
# 设置 rootless 权限
root@worker1$ hcu-ctk rootless

# test 用户运行容器，并挂载 test1 用户目录 
test@worker1$ docker run -ti --rm -e HCU_VISIBLE_DEVICES=0 -v /home/test1:/home/test1:rw <IMAGE> bash
# 无法在容器内的 test1 目录中创建新文件
root@5a3b63e3f91b:/workspace# touch /home/test1/aa
touch: cannot touch '/home/test1/aa': Read-only file system
```

## 4.6 查看占用 HCU 的容器

hcu-ctk container 子命令可查看占用 HCU 的容器，其可用参数如下：

- `--hcus`或`-d`：指定 HCU Index，默认值为`all`
- `--runtime`或`-r`：指定容器运行时，可选`docker`、`containerd`、`podman`，默认值为`docker`
- `--namespace`或`-n`：指定 Containerd 容器所在命名空间，默认值为`default`

> [!TIP]  
> 该功能需占用 HCU 的容器内有相应进程启动，若只是被挂载到容器而未被使用，则不会将对应容器返回  
> 若指定容器运行时为`containerd`，则需环境已安装 nerdctl 工具

示例用法如下：

```bash
# 查看占用 HCU 0 的 Docker 容器信息
$ hcu-ctk container -d 0
+--------+-------------------+------------------+
| HCU ID | Unique ID         | Container Names  |
+--------+-------------------+------------------+
| 0      | 0x7100F95CEE8450C1| magical_archimedes|
+--------+-------------------+------------------+
# 查看占用 HCU 且运行在 K8s.io 命名空间内的 Containerd 容器
$ hcu-ctk container -r containerd -n k8s.io
+--------+-------------------+----------------+
| HCU ID | Unique ID         | Container Names|
+--------+-------------------+----------------+
| 0      | 0x73873C7BC4E61081| custom-cc023   |
+--------+-------------------+----------------+
```
