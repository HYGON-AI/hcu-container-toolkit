# HCU Container Toolkit

## 简介

HCU Container Toolkit 是一个容器扩展工具套件，可以简化用户在容器中使用 HCU 设备的操作。该套件包括以下工具包。

- `hcu-container-runtime` - HCU 容器运行时
- `hcu-ctk` - HCU 容器工具集命令行

## 前置依赖

- 已安装容器运行时（Docker、Containerd 或 Podman）
- 已安装 HCU 驱动
- 需具备 root 权限来完成本工具套件的安装及配置写入

## 编译

编译方式分为两种，第一种可以依赖 docker 容器进行编译支持多种操作系统,只需在 make 后添加操作系统版本，例如

```bash
#编译rpm包
make rocky8

#编译deb包
make ubuntu22.04
```

由于 hy-cpu 对 gcc 的兼容性，增添了离线编译，会根据编译系统版本生成对应的 rpm 或 deb 格式的安装包。

```bash
bash build.sh
```

## 安装

使用 dpkg/rpm -i 进行安装。

```bash
# 对于Ubuntu、Debian等系统，使用dpkg安装
dpkg -i dist/hcu-container-toolkit_1.3.1-1_amd64.deb

# 对于Rocky、CentOS等系统，使用rpm安装
rpm -i dist/hcu-container-toolkit-1.3.1-1.x86_64.rpm
```

安装后会自动执行以下命令。

```bash
 # 生成配置文件
hcu-ctk --quiet config --config-file=/etc/hcu-container-runtime/config.toml --in-place
 # 修改docker的config.json的runtime
hcu-ctk runtime configure --runtime=docker --set-as-default 
```

安装完成后，若目标容器运行时为 Docker，则重启 Docker 服务。

```bash
systemctl restart docker
```

若需卸载 HCU Container Toolkit，则建议在卸载之前回滚容器运行时配置。
```bash
# 指定容器运行时回滚其配置
hcu-ctk runtime configure --runtime=docker --rollback
# 重启容器运行时服务
systemctl restart docker

# 卸载HCU Container Toolkit
# Ubuntu、Debian等系统
dpkg -P hcu-container-toolkit
# Rocky、CentOS等系统
rpm -e hcu-container-toolkit
```

## 使用说明

### 运行时集成

根据所用容器运行时，请移步相应的文档。

- [Docker](docs/user-guide/docker.md)
- [Containerd](docs/user-guide/containerd.md)
- [Podman](docs/user-guide/podman.md)

若需要在 k8s Device Plugin 中结合 HCU Container Toolkit 移除特权模式，可参考[此文档](docs/user-guide/k8s-device-plugin.md)

### 工具集用法

hcu-ctk 是进行运行时配置及管理 HCU 容器的命令行工具，可根据需求执行对应子命令。其各子命令使用可移步相应的文档。

- [runtime](docs/user-guide/hcu-ctk/runtime.md)
- [cdi](docs/user-guide/hcu-ctk/cdi.md)
- [config](docs/user-guide/hcu-ctk/config.md)
- [hcu-tracker](docs/user-guide/hcu-ctk/hcu-tracker.md)
- [rootless](docs/user-guide/hcu-ctk/rootless.md)
- [container](docs/user-guide/hcu-ctk/container.md)

## 许可证

本项目（HCU Container Toolkit）以 **Apache License 2.0** 许可证发布，详见根目录 [LICENSE](LICENSE) 文件。

项目中通过 `vendor/` 目录分发的第三方开源组件，按其各自许可证与来源在 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) 中逐项登记；使用本项目即表示同时同意相关第三方组件的许可证条款。
