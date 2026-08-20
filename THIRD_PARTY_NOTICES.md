# 第三方开源组件清单 (THIRD_PARTY_NOTICES)

> 生成日期：2026-08-20  
> 本仓库（HCU Container Toolkit）整体以 **Apache-2.0** 许可证发布（见根目录 `LICENSE`）。
> 以下为随仓库 `vendor/` 目录分发的第三方开源组件，按其真实许可证与来源逐项登记，
> 以遵守相应上游许可证与版权声明。组件均按上游原样引入，**HYGON 未做修改**（除非在“HYGON 修改”列特别说明）。

## 汇总

| # | 组件 | 版本 | 许可证 | 本地路径 |
|---|-------|------|--------|----------|
| 1 | `github.com/cpuguy83/go-md2man/v2` | v2.0.6 | MIT | `vendor/github.com/cpuguy83/go-md2man/v2` |
| 2 | `github.com/fsnotify/fsnotify` | v1.8.0 | BSD-3-Clause | `vendor/github.com/fsnotify/fsnotify` |
| 3 | `github.com/gofrs/flock` | v0.13.0 | BSD-3-Clause | `vendor/github.com/gofrs/flock` |
| 4 | `github.com/opencontainers/runtime-spec` | v1.2.1 | Apache-2.0 | `vendor/github.com/opencontainers/runtime-spec` |
| 5 | `github.com/opencontainers/runtime-tools` | v0.9.1-0.20221107090550-2e043c6bd626 | Apache-2.0 | `vendor/github.com/opencontainers/runtime-tools` |
| 6 | `github.com/pelletier/go-toml` | v1.9.5 | Apache-2.0 | `vendor/github.com/pelletier/go-toml` |
| 7 | `github.com/russross/blackfriday/v2` | v2.1.0 | BSD-2-Clause | `vendor/github.com/russross/blackfriday/v2` |
| 8 | `github.com/sirupsen/logrus` | v1.9.3 | MIT | `vendor/github.com/sirupsen/logrus` |
| 9 | `github.com/syndtr/gocapability` | v0.0.0-20200815063812-42c35b437635 | BSD-2-Clause | `vendor/github.com/syndtr/gocapability` |
| 10 | `github.com/urfave/cli/v2` | v2.27.5 | MIT | `vendor/github.com/urfave/cli/v2` |
| 11 | `github.com/xrash/smetrics` | v0.0.0-20240521201337-686a1a2994c1 | MIT | `vendor/github.com/xrash/smetrics` |
| 12 | `golang.org/x/mod` | v0.23.0 | BSD-3-Clause | `vendor/golang.org/x/mod` |
| 13 | `golang.org/x/sys` | v0.37.0 | BSD-3-Clause | `vendor/golang.org/x/sys` |
| 14 | `gopkg.in/ini.v1` | v1.67.0 | Apache-2.0 | `vendor/gopkg.in/ini.v1` |
| 15 | `sigs.k8s.io/yaml` | v1.4.0 | Apache-2.0 | `vendor/sigs.k8s.io/yaml` |
| 16 | `tags.cncf.io/container-device-interface` | v0.8.1 | Apache-2.0 | `vendor/tags.cncf.io/container-device-interface` |
| 17 | `tags.cncf.io/container-device-interface/specs-go` | v0.8.0 | Apache-2.0 | `vendor/tags.cncf.io/container-device-interface/specs-go` |

> 共登记 17 个 vendored 第三方组件；其许可证均为 HYGON 准入许可（MIT / BSD-2-Clause / BSD-3-Clause / Apache-2.0）。

## 逐项登记

### github.com/cpuguy83/go-md2man/v2

- **项目/仓库**：https://github.com/cpuguy83/go-md2man/v2
- **固定版本**：v2.0.6
- **许可证**：MIT
- **本地路径**：`vendor/github.com/cpuguy83/go-md2man/v2`
- **版权声明**：Copyright (c) 2014 Brian Goff; The above copyright notice and this permission notice shall be included in all; AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
- **HYGON 修改**：无（按上游原样 vendored）

### github.com/fsnotify/fsnotify

- **项目/仓库**：https://github.com/fsnotify/fsnotify
- **固定版本**：v1.8.0
- **许可证**：BSD-3-Clause
- **本地路径**：`vendor/github.com/fsnotify/fsnotify`
- **版权声明**：Copyright © 2012 The Go Authors. All rights reserved.; Copyright © fsnotify Authors. All rights reserved.; * Redistributions of source code must retain the above copyright notice, this
- **HYGON 修改**：无（按上游原样 vendored）

### github.com/gofrs/flock

- **项目/仓库**：https://github.com/gofrs/flock
- **固定版本**：v0.13.0
- **许可证**：BSD-3-Clause
- **本地路径**：`vendor/github.com/gofrs/flock`
- **已登记文件**：
  - `vendor/github.com/gofrs/flock/flock.go`
  - `vendor/github.com/gofrs/flock/flock_others.go`
  - `vendor/github.com/gofrs/flock/flock_unix.go`
  - `vendor/github.com/gofrs/flock/flock_unix_fcntl.go`
  - `vendor/github.com/gofrs/flock/flock_windows.go`
- **版权声明**：Copyright (c) 2018-2025, The Gofrs; Copyright (c) 2015-2020, Tim Heckman; * Redistributions of source code must retain the above copyright notice, this
- **HYGON 修改**：无（按上游原样 vendored）

### github.com/opencontainers/runtime-spec

- **项目/仓库**：https://github.com/opencontainers/runtime-spec
- **固定版本**：v1.2.1
- **许可证**：Apache-2.0
- **本地路径**：`vendor/github.com/opencontainers/runtime-spec`
- **版权声明**："Licensor" shall mean the copyright owner or entity authorized by; the copyright owner that is granting the License.; copyright notice that is included in or attached to the work
- **HYGON 修改**：无（按上游原样 vendored）

### github.com/opencontainers/runtime-tools

- **项目/仓库**：https://github.com/opencontainers/runtime-tools
- **固定版本**：v0.9.1-0.20221107090550-2e043c6bd626
- **许可证**：Apache-2.0
- **本地路径**：`vendor/github.com/opencontainers/runtime-tools`
- **版权声明**："Licensor" shall mean the copyright owner or entity authorized by; the copyright owner that is granting the License.; copyright notice that is included in or attached to the work
- **HYGON 修改**：无（按上游原样 vendored）

### github.com/pelletier/go-toml

- **项目/仓库**：https://github.com/pelletier/go-toml
- **固定版本**：v1.9.5
- **许可证**：Apache-2.0
- **本地路径**：`vendor/github.com/pelletier/go-toml`
- **版权声明**：Copyright (c) 2013 - 2021 Thomas Pelletier, Eric Anderton; The above copyright notice and this permission notice shall be included in all; AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
- **HYGON 修改**：无（按上游原样 vendored）

### github.com/russross/blackfriday/v2

- **项目/仓库**：https://github.com/russross/blackfriday/v2
- **固定版本**：v2.1.0
- **许可证**：BSD-2-Clause
- **本地路径**：`vendor/github.com/russross/blackfriday/v2`
- **版权声明**：> Copyright © 2011 Russ Ross; > 1.  Redistributions of source code must retain the above copyright; >     copyright notice, this list of conditions and the following
- **HYGON 修改**：无（按上游原样 vendored）

### github.com/sirupsen/logrus

- **项目/仓库**：https://github.com/sirupsen/logrus
- **固定版本**：v1.9.3
- **许可证**：MIT
- **本地路径**：`vendor/github.com/sirupsen/logrus`
- **版权声明**：Copyright (c) 2014 Simon Eskildsen; The above copyright notice and this permission notice shall be included in; AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
- **HYGON 修改**：无（按上游原样 vendored）

### github.com/syndtr/gocapability

- **项目/仓库**：https://github.com/syndtr/gocapability
- **固定版本**：v0.0.0-20200815063812-42c35b437635
- **许可证**：BSD-2-Clause
- **本地路径**：`vendor/github.com/syndtr/gocapability`
- **版权声明**：Copyright 2013 Suryandaru Triandana <syndtr@gmail.com>; * Redistributions of source code must retain the above copyright; * Redistributions in binary form must reproduce the above copyright
- **HYGON 修改**：无（按上游原样 vendored）

### github.com/urfave/cli/v2

- **项目/仓库**：https://github.com/urfave/cli/v2
- **固定版本**：v2.27.5
- **许可证**：MIT
- **本地路径**：`vendor/github.com/urfave/cli/v2`
- **版权声明**：Copyright (c) 2022 urfave/cli maintainers; The above copyright notice and this permission notice shall be included in all; AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
- **HYGON 修改**：无（按上游原样 vendored）

### github.com/xrash/smetrics

- **项目/仓库**：https://github.com/xrash/smetrics
- **固定版本**：v0.0.0-20240521201337-686a1a2994c1
- **许可证**：MIT
- **本地路径**：`vendor/github.com/xrash/smetrics`
- **版权声明**：Copyright (C) 2016 Felipe da Cunha Gonçalves; The above copyright notice and this permission notice shall be included in all; COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
- **HYGON 修改**：无（按上游原样 vendored）

### golang.org/x/mod

- **项目/仓库**：https://golang.org/x/mod
- **固定版本**：v0.23.0
- **许可证**：BSD-3-Clause
- **本地路径**：`vendor/golang.org/x/mod`
- **版权声明**：Copyright 2009 The Go Authors.; * Redistributions of source code must retain the above copyright; copyright notice, this list of conditions and the following disclaimer
- **HYGON 修改**：无（按上游原样 vendored）

### golang.org/x/sys

- **项目/仓库**：https://golang.org/x/sys
- **固定版本**：v0.37.0
- **许可证**：BSD-3-Clause
- **本地路径**：`vendor/golang.org/x/sys`
- **版权声明**：Copyright 2009 The Go Authors.; * Redistributions of source code must retain the above copyright; copyright notice, this list of conditions and the following disclaimer
- **HYGON 修改**：无（按上游原样 vendored）

### gopkg.in/ini.v1

- **项目/仓库**：https://gopkg.in/ini.v1
- **固定版本**：v1.67.0
- **许可证**：Apache-2.0
- **本地路径**：`vendor/gopkg.in/ini.v1`
- **版权声明**："Licensor" shall mean the copyright owner or entity authorized by the copyright; available under the License, as indicated by a copyright notice that is included; by the copyright owner or by an individual or Legal Entity authorized to submit
- **HYGON 修改**：无（按上游原样 vendored）

### sigs.k8s.io/yaml

- **项目/仓库**：https://github.com/kubernetes-sigs/yaml
- **固定版本**：v1.4.0
- **许可证**：Apache-2.0
- **本地路径**：`vendor/sigs.k8s.io/yaml`
- **版权声明**：Copyright (c) 2014 Sam Ghods; The above copyright notice and this permission notice shall be included in all; AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
- **HYGON 修改**：无（按上游原样 vendored）

### tags.cncf.io/container-device-interface

- **项目/仓库**：https://github.com/cncf-tags/container-device-interface
- **固定版本**：v0.8.1
- **许可证**：Apache-2.0
- **本地路径**：`vendor/tags.cncf.io/container-device-interface`
- **版权声明**："Licensor" shall mean the copyright owner or entity authorized by; the copyright owner that is granting the License.; copyright notice that is included in or attached to the work
- **HYGON 修改**：无（按上游原样 vendored）

### tags.cncf.io/container-device-interface/specs-go

- **项目/仓库**：https://github.com/cncf-tags/container-device-interface/specs-go
- **固定版本**：v0.8.0
- **许可证**：Apache-2.0
- **本地路径**：`vendor/tags.cncf.io/container-device-interface/specs-go`
- **版权声明**："Licensor" shall mean the copyright owner or entity authorized by; the copyright owner that is granting the License.; copyright notice that is included in or attached to the work
- **HYGON 修改**：无（按上游原样 vendored）

## 已登记文件（完整路径索引）

> 以下为随仓库 `vendor/` 目录分发的第三方源码文件完整相对路径索引，供合规扫描按路径核对来源清单登记。
> 各文件所属项目的仓库、固定版本、许可证与版权声明见上文对应组件条目；生成文件见 `*_mock.go` 说明。

### vendor/github.com/gofrs/flock

- `vendor/github.com/gofrs/flock/flock.go`
- `vendor/github.com/gofrs/flock/flock_others.go`
- `vendor/github.com/gofrs/flock/flock_unix.go`
- `vendor/github.com/gofrs/flock/flock_unix_fcntl.go`
- `vendor/github.com/gofrs/flock/flock_windows.go`

### vendor/github.com/pelletier/go-toml

- `vendor/github.com/pelletier/go-toml/localtime.go`

### vendor/github.com/russross/blackfriday/v2

- `vendor/github.com/russross/blackfriday/v2/block.go`
- `vendor/github.com/russross/blackfriday/v2/html.go`
- `vendor/github.com/russross/blackfriday/v2/inline.go`
- `vendor/github.com/russross/blackfriday/v2/markdown.go`
- `vendor/github.com/russross/blackfriday/v2/smartypants.go`

### vendor/github.com/sirupsen/logrus

- `vendor/github.com/sirupsen/logrus/alt_exit.go`

### vendor/github.com/syndtr/gocapability

- `vendor/github.com/syndtr/gocapability/capability/capability.go`
- `vendor/github.com/syndtr/gocapability/capability/capability_linux.go`
- `vendor/github.com/syndtr/gocapability/capability/capability_noop.go`
- `vendor/github.com/syndtr/gocapability/capability/enum.go`
- `vendor/github.com/syndtr/gocapability/capability/syscall_linux.go`

### vendor/github.com/urfave/cli/v2

- `vendor/github.com/urfave/cli/v2/app.go`
- `vendor/github.com/urfave/cli/v2/template.go`

### vendor/golang.org/x/mod

- `vendor/golang.org/x/mod/semver/semver.go`

### vendor/golang.org/x/sys

- `vendor/golang.org/x/sys/unix/affinity_linux.go`
- `vendor/golang.org/x/sys/unix/aliases.go`
- `vendor/golang.org/x/sys/unix/auxv.go`
- `vendor/golang.org/x/sys/unix/auxv_unsupported.go`
- `vendor/golang.org/x/sys/unix/bluetooth_linux.go`
- `vendor/golang.org/x/sys/unix/bpxsvc_zos.go`
- `vendor/golang.org/x/sys/unix/cap_freebsd.go`
- `vendor/golang.org/x/sys/unix/constants.go`
- `vendor/golang.org/x/sys/unix/dev_aix_ppc.go`
- `vendor/golang.org/x/sys/unix/dev_aix_ppc64.go`
- `vendor/golang.org/x/sys/unix/dev_darwin.go`
- `vendor/golang.org/x/sys/unix/dev_dragonfly.go`
- `vendor/golang.org/x/sys/unix/dev_freebsd.go`
- `vendor/golang.org/x/sys/unix/dev_linux.go`
- `vendor/golang.org/x/sys/unix/dev_netbsd.go`
- `vendor/golang.org/x/sys/unix/dev_openbsd.go`
- `vendor/golang.org/x/sys/unix/dev_zos.go`
- `vendor/golang.org/x/sys/unix/dirent.go`
- `vendor/golang.org/x/sys/unix/endian_big.go`
- `vendor/golang.org/x/sys/unix/endian_little.go`
- `vendor/golang.org/x/sys/unix/env_unix.go`
- `vendor/golang.org/x/sys/unix/fcntl.go`
- `vendor/golang.org/x/sys/unix/fcntl_darwin.go`
- `vendor/golang.org/x/sys/unix/fcntl_linux_32bit.go`
- `vendor/golang.org/x/sys/unix/fdset.go`
- `vendor/golang.org/x/sys/unix/gccgo.go`
- `vendor/golang.org/x/sys/unix/gccgo_c.c`
- `vendor/golang.org/x/sys/unix/gccgo_linux_amd64.go`
- `vendor/golang.org/x/sys/unix/ifreq_linux.go`
- `vendor/golang.org/x/sys/unix/ioctl_linux.go`
- `vendor/golang.org/x/sys/unix/ioctl_signed.go`
- `vendor/golang.org/x/sys/unix/ioctl_unsigned.go`
- `vendor/golang.org/x/sys/unix/ioctl_zos.go`
- `vendor/golang.org/x/sys/unix/mkall.sh`
- `vendor/golang.org/x/sys/unix/mkerrors.sh`
- `vendor/golang.org/x/sys/unix/mmap_nomremap.go`
- `vendor/golang.org/x/sys/unix/mremap.go`
- `vendor/golang.org/x/sys/unix/pagesize_unix.go`
- `vendor/golang.org/x/sys/unix/pledge_openbsd.go`
- `vendor/golang.org/x/sys/unix/ptrace_darwin.go`
- `vendor/golang.org/x/sys/unix/ptrace_ios.go`
- `vendor/golang.org/x/sys/unix/race.go`
- `vendor/golang.org/x/sys/unix/race0.go`
- `vendor/golang.org/x/sys/unix/readdirent_getdents.go`
- `vendor/golang.org/x/sys/unix/readdirent_getdirentries.go`
- `vendor/golang.org/x/sys/unix/sockcmsg_dragonfly.go`
- `vendor/golang.org/x/sys/unix/sockcmsg_linux.go`
- `vendor/golang.org/x/sys/unix/sockcmsg_unix.go`
- `vendor/golang.org/x/sys/unix/sockcmsg_unix_other.go`
- `vendor/golang.org/x/sys/unix/sockcmsg_zos.go`
- `vendor/golang.org/x/sys/unix/syscall.go`
- `vendor/golang.org/x/sys/unix/syscall_aix.go`
- `vendor/golang.org/x/sys/unix/syscall_aix_ppc.go`
- `vendor/golang.org/x/sys/unix/syscall_aix_ppc64.go`
- `vendor/golang.org/x/sys/unix/syscall_bsd.go`
- `vendor/golang.org/x/sys/unix/syscall_darwin.go`
- `vendor/golang.org/x/sys/unix/syscall_darwin_amd64.go`
- `vendor/golang.org/x/sys/unix/syscall_darwin_arm64.go`
- `vendor/golang.org/x/sys/unix/syscall_darwin_libSystem.go`
- `vendor/golang.org/x/sys/unix/syscall_dragonfly.go`
- `vendor/golang.org/x/sys/unix/syscall_dragonfly_amd64.go`
- `vendor/golang.org/x/sys/unix/syscall_freebsd.go`
- `vendor/golang.org/x/sys/unix/syscall_freebsd_386.go`
- `vendor/golang.org/x/sys/unix/syscall_freebsd_amd64.go`
- `vendor/golang.org/x/sys/unix/syscall_freebsd_arm.go`
- `vendor/golang.org/x/sys/unix/syscall_freebsd_arm64.go`
- `vendor/golang.org/x/sys/unix/syscall_freebsd_riscv64.go`
- `vendor/golang.org/x/sys/unix/syscall_hurd.go`
- `vendor/golang.org/x/sys/unix/syscall_hurd_386.go`
- `vendor/golang.org/x/sys/unix/syscall_illumos.go`
- `vendor/golang.org/x/sys/unix/syscall_linux.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_386.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_alarm.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_amd64.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_amd64_gc.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_arm.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_arm64.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_gc.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_gc_386.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_gc_arm.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_gccgo_386.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_gccgo_arm.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_loong64.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_mips64x.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_mipsx.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_ppc.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_ppc64x.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_riscv64.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_s390x.go`
- `vendor/golang.org/x/sys/unix/syscall_linux_sparc64.go`
- `vendor/golang.org/x/sys/unix/syscall_netbsd.go`
- `vendor/golang.org/x/sys/unix/syscall_netbsd_386.go`
- `vendor/golang.org/x/sys/unix/syscall_netbsd_amd64.go`
- `vendor/golang.org/x/sys/unix/syscall_netbsd_arm.go`
- `vendor/golang.org/x/sys/unix/syscall_netbsd_arm64.go`
- `vendor/golang.org/x/sys/unix/syscall_openbsd.go`
- `vendor/golang.org/x/sys/unix/syscall_openbsd_386.go`
- `vendor/golang.org/x/sys/unix/syscall_openbsd_amd64.go`
- `vendor/golang.org/x/sys/unix/syscall_openbsd_arm.go`
- `vendor/golang.org/x/sys/unix/syscall_openbsd_arm64.go`
- `vendor/golang.org/x/sys/unix/syscall_openbsd_libc.go`
- `vendor/golang.org/x/sys/unix/syscall_openbsd_mips64.go`
- `vendor/golang.org/x/sys/unix/syscall_openbsd_ppc64.go`
- `vendor/golang.org/x/sys/unix/syscall_openbsd_riscv64.go`
- `vendor/golang.org/x/sys/unix/syscall_solaris.go`
- `vendor/golang.org/x/sys/unix/syscall_solaris_amd64.go`
- `vendor/golang.org/x/sys/unix/syscall_unix.go`
- `vendor/golang.org/x/sys/unix/syscall_unix_gc.go`
- `vendor/golang.org/x/sys/unix/syscall_unix_gc_ppc64x.go`
- `vendor/golang.org/x/sys/unix/syscall_zos_s390x.go`
- `vendor/golang.org/x/sys/unix/sysvshm_linux.go`
- `vendor/golang.org/x/sys/unix/sysvshm_unix.go`
- `vendor/golang.org/x/sys/unix/sysvshm_unix_other.go`
- `vendor/golang.org/x/sys/unix/timestruct.go`
- `vendor/golang.org/x/sys/unix/unveil_openbsd.go`
- `vendor/golang.org/x/sys/unix/vgetrandom_linux.go`
- `vendor/golang.org/x/sys/unix/vgetrandom_unsupported.go`
- `vendor/golang.org/x/sys/unix/xattr_bsd.go`
- `vendor/golang.org/x/sys/unix/zerrors_zos_s390x.go`
- `vendor/golang.org/x/sys/unix/ztypes_zos_s390x.go`
- `vendor/golang.org/x/sys/windows/aliases.go`
- `vendor/golang.org/x/sys/windows/dll_windows.go`
- `vendor/golang.org/x/sys/windows/env_windows.go`
- `vendor/golang.org/x/sys/windows/eventlog.go`
- `vendor/golang.org/x/sys/windows/exec_windows.go`
- `vendor/golang.org/x/sys/windows/memory_windows.go`
- `vendor/golang.org/x/sys/windows/mksyscall.go`
- `vendor/golang.org/x/sys/windows/race.go`
- `vendor/golang.org/x/sys/windows/race0.go`
- `vendor/golang.org/x/sys/windows/security_windows.go`
- `vendor/golang.org/x/sys/windows/service.go`
- `vendor/golang.org/x/sys/windows/setupapi_windows.go`
- `vendor/golang.org/x/sys/windows/str.go`
- `vendor/golang.org/x/sys/windows/syscall.go`
- `vendor/golang.org/x/sys/windows/syscall_windows.go`
- `vendor/golang.org/x/sys/windows/types_windows.go`
- `vendor/golang.org/x/sys/windows/types_windows_386.go`
- `vendor/golang.org/x/sys/windows/types_windows_amd64.go`
- `vendor/golang.org/x/sys/windows/types_windows_arm.go`
- `vendor/golang.org/x/sys/windows/types_windows_arm64.go`

### vendor/gopkg.in/ini.v1

- `vendor/gopkg.in/ini.v1/data_source.go`
- `vendor/gopkg.in/ini.v1/deprecated.go`
- `vendor/gopkg.in/ini.v1/error.go`
- `vendor/gopkg.in/ini.v1/file.go`
- `vendor/gopkg.in/ini.v1/helper.go`
- `vendor/gopkg.in/ini.v1/ini.go`
- `vendor/gopkg.in/ini.v1/key.go`
- `vendor/gopkg.in/ini.v1/parser.go`
- `vendor/gopkg.in/ini.v1/section.go`
- `vendor/gopkg.in/ini.v1/struct.go`

### vendor/sigs.k8s.io/yaml

- `vendor/sigs.k8s.io/yaml/fields.go`
- `vendor/sigs.k8s.io/yaml/yaml.go`
- `vendor/sigs.k8s.io/yaml/yaml_go110.go`

### vendor/tags.cncf.io/container-device-interface

- `vendor/tags.cncf.io/container-device-interface/internal/validation/k8s/objectmeta.go`
- `vendor/tags.cncf.io/container-device-interface/internal/validation/k8s/validation.go`
- `vendor/tags.cncf.io/container-device-interface/internal/validation/validate.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/annotations.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/cache.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/cache_test_unix.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/cache_test_windows.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/container-edits.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/container-edits_unix.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/container-edits_windows.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/default-cache.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/device.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/oci.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/qualified-device.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/registry.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/spec-dirs.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/spec.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/spec_linux.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/spec_other.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/cdi/version.go`
- `vendor/tags.cncf.io/container-device-interface/pkg/parser/parser.go`

### vendor/tags.cncf.io/container-device-interface/specs-go

- `vendor/tags.cncf.io/container-device-interface/specs-go/oci.go`
## 说明

- 各组件完整的许可证全文见其 `vendor/<组件>/LICENSE*` 文件，需一并分发。
- `go.mod` 中显式声明但未被实际引入、因而未进入 `vendor/` 树的依赖：
  - `github.com/kr/text`（v0.2.0）—— 未 vendored，不在分发范围内。
  - `github.com/rogpeppe/go-internal`（v1.14.1）—— 未 vendored，不在分发范围内。
- 生成文件（如 `*_mock.go`）来源于对应 mock 生成器，按上游原样引入，不直接修改生成产物。
