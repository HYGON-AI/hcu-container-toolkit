# 运行时配置调整

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

# 通过dry-run在控制台中查看待写入的配置
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

# 移除Docker配置文件中HCU Runtime部分
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

# 回滚Docker配置文件
$ hcu-ctk runtime configure --runtime=docker --rollback
WARN[0000] Ignoring runtime-config-override flag for docker 
INFO[0000] Restored the backup file /etc/docker/daemon.json.hcu-ctk.bak 
$ cat /etc/docker/daemon.json 
{
  "debug": false
}

```
