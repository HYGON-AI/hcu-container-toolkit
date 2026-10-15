# 2 HCU-Container-Toolkit 安装

## 2.1 前置依赖

- 已安装容器运行时（Docker、Containerd 或 Podman）
- 已安装 HCU 驱动
- 需具备 root 权限来完成本工具套件的安装及配置写入

## 2.2 安装

从社区获取[安装软件包](https://download.sourcefind.cn:65024/5/main/)，使用 dpkg/rpm -i 进行安装。

```bash
# 对于 Ubuntu、Debian 等系统，使用 dpkg 安装
dpkg -i hcu-container-toolkit_1.3.1-1_amd64.deb

# 对于 Rocky、CentOS 等系统，使用 rpm 安装
rpm -i hcu-container-toolkit-1.3.1-1.x86_64.rpm
```

安装后会自动执行以下命令。

```bash
 # 生成配置文件
hcu-ctk --quiet config --config-file=/etc/hcu-container-runtime/config.toml --in-place
 # 修改 docker 的 config.json 的 runtime
hcu-ctk runtime configure --runtime=docker --set-as-default 
```

安装完成后，若目标容器运行时为 Docker，则重启 Docker 服务。

```bash
systemctl daemon-reload
systemctl restart docker
```

若需卸载 HCU Container Toolkit，则建议在卸载之前回滚容器运行时配置。

```bash
# 指定容器运行时回滚其配置
hcu-ctk runtime configure --runtime=docker --rollback
# 重启容器运行时服务
systemctl daemon-reload
systemctl restart docker

# 卸载 HCU Container Toolkit
# Ubuntu、Debian 等系统
dpkg -P hcu-container-toolkit
# Rocky、CentOS 等系统
rpm -e hcu-container-toolkit
```
