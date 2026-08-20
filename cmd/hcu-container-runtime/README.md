# HCU Container Runtime

HCU Container Runtime 是一个专为符合 OCI （Open Container Initiative，开放容器计划）规范的底层运行时（ 如[runc](https://github.com/opencontainers/runc) ）设计的垫片（shim）。当 hcu-container-runtime 接收到一个 `create` 命令时，它会对传入的 [OCI运行时规范](https://github.com/opencontainers/runtime-spec) 进行即时修改，以添加对 HCU 的特殊支持，然后将这个修改后的命令转发给底层的运行时（如runc、containerd等）。

## 配置

HCU Container Runtime （HCR） 使用基于文件的配置方式，其配置文件存储在 /etc/hcu-container-runtime/config.toml 路径下。安装 hcu-container-toolkit 时会自动生成该文件。
