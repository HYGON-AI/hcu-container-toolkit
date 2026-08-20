# HCU 可访问性配置

HCU Tracker 可自行维护 HCU 状态及容器对 HCU 的可访问性配置。默认情况下 HCU Tracker 是 disable 状态，用户可自行配置启用。

> [!TIP]
> 目前支持通过`--gpus`和`-e HCU_VISIBLE_DEVICES`方式启动的容器


## 启用/禁用 HCU Tracker

可通过 hcu-ctk 对 HCU Tracker 的启用或禁用进行调整。

```bash
# 功能启用
hcu-ctk hcu-tracker enable
# 功能禁用
hcu-ctk hcu-tracker disable
```

## 查看 HCU Tracker Status

在 HCU Tracker 功能启用后，可通过 status 查询 HCU 可访问性及对应容器的情况。

```bash
hcu-ctk hcu-tracker status
------------------------------------------------------------------------------------------------------------------------
HCU Id    UUID                     Accessibility       Container Ids                                                    
------------------------------------------------------------------------------------------------------------------------
0         0x7100F95F9B3230E1       Shared              b3277bc4c47b550d91ab97f40aa66e881a32e031a988d1443997ab3aea3fafa2 
1         0x7100F95F9B102101       Shared              None 

```

## 调整容器对 HCU 的访问权限

HCU Tracker 中 HCU 的可访问性有两种：
- `shared` 表示 HCU 可以同时被多个容器访问。默认情况下，所有 HCU 都被授予 shared 访问权限，以符合 Docker 的默认行为。
- `exclusive` 表示 HCU 在任何时刻最多只能被一个容器访问。

可通过如下命令调整权限。

```bash
# 调整多个HCU为exclusive权限
# 若HCU此时已被多个容器使用，则配置失败
hcu-ctk hcu-tracker 0-3 exclusive

# 调整单个HCU为exclusive权限
hcu-ctk hcu-tracker 4 exclusive

# 调整HCU为shared权限
hcu-ctk hcu-tracker 4 shared
```
## 重置 HCU Tracker

在 HCU 进行虚拟化操作后，应重置 HCU Tracker。

```bash
hcu-ctk hcu-tracker reset
```
