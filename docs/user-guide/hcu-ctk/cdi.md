# CDI 管理

hcu-ctk cdi 子命令可以实现 CDI 规范生成、校验、展示等功能，以便在容器中访问和使用 CDI 设备。

> [!TIP]
> CDI 介绍可参考[此地址](https://github.com/cncf-tags/container-device-interface)


## cdi list

hcu-ctk cdi list 子命令可通过已生成的 CDI 规范展示相应 CDI 设备。该子命令提供了`--spec-dir`参数指定 CDI 规范文件所在目录，参数默认值为`“/etc/cdi,/var/run/cdi”`。

示例用法如下：

```bash
# 从默认目录读取CDI规范文件
$ hcu-ctk cdi list
INFO[0000] Found 1 CDI devices                          
hygon.com/hcu=0
hygon.com/hcu=all
hygon.com/hcu=hcu-TPXS300002100601

# 指定文件目录读取
# hcu-ctk cdi list --spec-dir /tmp
INFO[0000] Found 1 CDI devices                          
hygon.com/hcu=0
hygon.com/hcu=all
hygon.com/hcu=hcu-TPXS300002100601
```


## cdi generate

hcu-ctk cdi generate 子命令负责生成 CDI 规范文件，其可用参数如下：

- `--output`：CDI 规范写入文件路径，若设置空值或忽略则将 CDI 规范输出到控制台
- `--format`：CDI 规范的格式，可配置为`json`或`yaml`，默认值为`yaml`
- `--device-name-strategy`：指定设备名称的生成策略，可选`index`、`uuid`、`type-index`，默认值为`index,uuid`
- `--hcu-cdi-hook-path`或`--hcu-ctk-path`：指定 hcu-cdi-hook 二进制的路径，一般情况下无需指定
- `--vendor`或`--cdi-vendor`：自定义设备商 ID，默认值为`hygon.cn`
- `--class`或`--cdi-class`：自定义设备类型，默认值为`hcu`

示例用法如下：

```bash
# 生成CDI规范，并写入到/etc/cdi/hcu.yaml中
$ hcu-ctk cdi generate --output /etc/cdi/hcu.yaml
$ hcu-ctk cdi list
INFO[0000] Found 1 CDI devices                          
hygon.com/hcu=0
hygon.com/hcu=all
hygon.com/hcu=hcu-TPXS300002100601

# 指定设备名称生成策略为type-index和index
$ hcu-ctk cdi generate --device-name-strategy "type-index,index" --output /etc/cdi/hcu.yaml
$ hcu-ctk cdi list
INFO[0000] Found 1 CDI devices                          
hygon.com/hcu=0
hygon.com/hcu=all
hygon.com/hcu=hcu0
```


## cdi validate

hcu-ctk cdi validate 子命令可对已生成的 CDI 规范文件进行校验，该功能适用于 HCU 硬件环境有变动的场景。该子命令提供了`--path`参数指定 CDI 规范文件路径，参数默认值为`/etc/cdi/hcu.yaml`。

示例用法如下：

```bash
# 校验默认路径的CDI规范文件
$ hcu-ctk cdi validate
...
CDI spec is valid

# 指定CDI规范文件路径
$ hcu-ctk cdi validate --path /tmp/hcu.yaml
...
CDI spec is valid
```

## cdi transform 

hcu-ctk cdi transform root 子命令可对 CDI 规范中挂载目录进行转换，该功能适用于驱动目录等有调整的场景。其可用参数如下：

- `--from`：指定待转换目录
- `--to`：指定转换后的目录
- `--relative-to`：指定转换的目录为本地目录或容器内目录，可选`host`和`container`，默认值为`host`
- `--input`：指定待转换文件，也可指定为`-`，代表在控制台输入 CDI 规范。默认值为`-`
- `--output`：指定转换后写入的文件，若未指定或指定为空，则将 CDI 规范输出到控制台

示例用法如下：

```bash
# 将/etc/cdi/hcu.yaml中挂载信息的本地目录/opt/hyhal调整为/usr/local/hyhal，输出到控制台
hcu-ctk cdi transform root --from /opt/hyhal --to /usr/local/hyhal --relative-to host --input /etc/cdi/hcu.yaml

# 将/etc/cdi/hcu.yaml中挂载信息的本地目录/opt/hyhal调整为/usr/local/hyhal，并写回到原文件
hcu-ctk cdi transform root --from /opt/hyhal --to /usr/local/hyhal --relative-to host --input /etc/cdi/hcu.yaml --output /etc/cdi/hcu.yaml
```
