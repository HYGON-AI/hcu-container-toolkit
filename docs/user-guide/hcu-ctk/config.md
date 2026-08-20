# HCU Container Toolkit 配置管理

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

# 从指定文件获取配置并进行调整，并将结果写入到/tmp/config.toml中
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
