# HCU Container Toolkit Changelog

## v1.4.0 - 2026-11-19

### Added
- 支持通过环境变量动态切分并挂载虚拟卡
- 新增`hcu-ctk vhcu`子命令，提供动态切分环境自检（check）、容量与归属查询（status/list/show）及兜底回收（gc）等功能
- 新增`hcu-container-runtime.vhcu`配置节，支持调整动态切分开关、每卡虚拟卡数量上限等参数

### Fixed
- 修复主机存在多套 DTK（如`/opt/hyhal`与`/opt/rocm`并存）时，容器运行时可能解析到非预期 hy-smi、导致动态切分创建失败的问题

## v1.3.1 - 2026-07-22

### Added
- 支持对 Containerd 2+ 版本 version 3 配置调整

## v1.3.0 - 2026-03-12

### Added
- 支持 HCU Device Plugin 移除特权模式运行
- 支持通过 MIG_VISIBLE_DEVICES 挂载 MIG 设备
- 支持移除默认运行时配置
- 支持移除 HCU Runtime 配置
- 支持回滚容器运行时配置
- hcu-ctk container 新增支持 Container 和 Podman 容器

### Fixed
- 修复因系统语言设置导致的编译错误
- 修复 Podman 配置调整 dry-run 无法使用的问题
- 修复 hcu-ctk container 子命令容器名称不显示的问题
- 调整 hcu-ctk cdi validate 逻辑，修复特定条件下的报错
- 修复 hcu-ctk config 配置默认值不生效、参数校验不生效的问题
- 修复 hcu-ctk config --config-file 无法保留已修改配置的问题
- 修复调用 smi 工具获取不到 HCU Index 的问题

### Changed
- 优化 Podman 配置调整及文件写入位置
- hook 移除 hyhal 目录链接，以避免 DP 动态分配 vHCU 时删除 hyhal 目录下文件
- 优化编译脚本中对 GO 环境的处理
- 调整 hcu-ctk rootless 可配置容器运行时为 Docker
- 调整 hcu-ctk docker 子命令为 container
- 规范 CDI 名称、资源名称、命令行工具参数等相关用词
- 优化 JSON 格式的 CDI 规范文件的可读性

## v1.2.3
bugfix：
- 去掉默认的 ApparmorProfile，让挂载的 HCU 设备系统文件有可写权限。

## v1.2.2
RDMA：设置环境变量 `HCU_MOFED` 为 `enabled`时，开启 RDMA 支持。
- 挂载发现的 MOFED Infiniband 设备
- --cap-add=SYS_LOCK

## v1.2.1
xprof 支持
- 挂载设备 /dev/mem
- 挂载 HCU 设备对应的 PCI 系统文件 /sys/bus/pci/devices/${PCIBusId}
- --cap-add=SYS_RAWIO

## v1.2.0
- 添加 CDI 支持

## v1.1.1
- 安装时自动更新 docker 的运行时
- 给容器添加/opt/hyhal 链接

## v1.1.0
- initialize version
