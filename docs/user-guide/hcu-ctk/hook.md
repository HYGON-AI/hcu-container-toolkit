# 在容器规范中注入特定 hook 操作

> [!TIP]
> 该子命令仅用于 Docker 容器运行时

在 Docker 默认的 rootful 模式下，Linux 普通用户启动的容器在使用`-v`挂载目录时可以赋予目录写入权限，即使该目录是其他 Linux 用户目录也可在容器内对目录下文件进行删增修改等操作。为避免这一问题，可执行`hcu-ctk rootless`命令进行限制。

示例用法如下：

```bash
# 设置rootless权限
root@worker1$ hcu-ctk rootless

# test用户运行容器，并挂载test1用户目录 
test@worker1$ docker run -ti --rm -e HCU_VISIBLE_DEVICES=0 -v /home/test1:/home/test1:rw <IMAGE> bash
# 无法在容器内的test1目录中创建新文件
root@5a3b63e3f91b:/workspace# touch /home/test1/aa
touch: cannot touch '/home/test1/aa': Read-only file system
```
