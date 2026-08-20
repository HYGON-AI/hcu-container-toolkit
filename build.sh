#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2026 Hygon Information Technology Co., Ltd.
#

set -e

GO_VERSION="go1.24.13"
GO_TAR="$GO_VERSION.linux-amd64.tar.gz"
GO_URL="https://golang.google.cn/dl/$GO_TAR"

check_go_installed() {
    if command -v go &>/dev/null; then
        echo "Go 环境已安装: $(go version)"
        return 0
    else
        echo "Go 环境未安装"
        return 1
    fi
}

prompt_download_go() {
    echo "Go安装包下载失败。请手动下载 $GO_VERSION 安装包并放到脚本相同目录"
    echo "下载链接: $GO_URL"
    echo "安装包准备完成后，请重新运行脚本"
    exit 1
}

download_go() {
    if [ ! -f "$GO_TAR" ]; then
        echo "未找到 $GO_TAR 文件，开始自动下载 Go 安装包..."
        wget "$GO_URL" -O "$GO_TAR" || { prompt_download_go; }
    fi
}

install_go() {
    echo "开始解压并安装 Go..."
    if [ -d "$HOME/go" ]; then
        echo "$HOME/go 目录已存在，备份并删除..."
        if [ -d "$HOME/go.bak" ]; then
            echo "$HOME/go.bak 已存在，删除旧备份..."
            rm -rf "$HOME/go.bak"
        fi
        mv "$HOME/go" "$HOME/go.bak"
    fi
    tar -C "$HOME" -xzf "$GO_TAR" || { echo "解压 Go 安装包失败"; exit 1; }
    export PATH="$HOME/go/bin:$PATH"
    echo "Go 安装成功"
}

# 从versions.mk读取版本信息
OS_NAME=$(grep '^NAME=' /etc/os-release | cut -d= -f2 | tr -d '"')
PACKAGE_TYPE="rpm"
case "$OS_NAME" in
    "Red Hat Enterprise Linux"*|"CentOS"*|"Rocky Linux"*|"AlmaLinux"*|"Oracle Linux"*|"Amazon Linux"*|"Fedora"*)
        PACKAGE_TYPE="rpm"
        ;;
    "Debian"*|"Ubuntu"*|"Linux Mint"*|"Kali"*|"Raspbian"*|"Elementary"*|"Pop!_OS"*|"Zorin"*)
        PACKAGE_TYPE="deb"
        ;;
    *)
        echo "Error: 不支持的操作系统 $OS_NAME" >&2
        exit 1
        ;;
esac

PACKAGE_VERSION=$(grep '^LIB_VERSION :=' versions.mk | awk '{print $3}')
PACKAGE_REVISION=$(grep '^PACKAGE_REVISION :=' versions.mk | awk '{print $3}')
PKG_NAME=$(grep '^LIB_NAME :=' versions.mk | awk '{print $3}')
if [ -z "$PACKAGE_VERSION" ] || [ -z "$PACKAGE_REVISION" ]; then
    echo "Error: 无法从versions.mk读取版本信息" >&2
    exit 1
fi

if ! check_go_installed; then
    download_go
    install_go
fi

make cmds
GIT_COMMIT=$(git describe --match="" --dirty --long --always --abbrev=40 2> /dev/null || echo "")
PKG_VERS="$PACKAGE_VERSION"
PKG_REV="$PACKAGE_REVISION"
if [ "$PACKAGE_TYPE" = "rpm" ]; then
    sources=("hcu-container-runtime" "hcu-cdi-hook" "hcu-ctk" "hcu-docker")
    dirs=("SOURCES" "RPMS" "BUILD" "SRPMS" "BUILDROOT")
    for dir in "${dirs[@]}"
    do
       if [ ! -d "$dir" ];then
           mkdir "$dir"
       fi
    done
    for file in "${sources[@]}"
    do
       rm -rf "SOURCES/$file" > /dev/null 2>&1
       cp -r "$file" "SOURCES/$file"
    done
    cp -rf packaging/rpm/SOURCES/* SOURCES/ > /dev/null 2>&1
    # 使用现代命令替换语法，提高可读性


    if [ ! -d dist ];then
       mkdir dist
    fi
    arch=$(uname -m)
    rpmbuild --clean --target="$arch" -bb \
             -D "_topdir $PWD" \
             -D "release_date $(LC_ALL=en_US.UTF-8 date +'%a %b %d %Y')" \
             -D "git_commit ${GIT_COMMIT}" \
             -D "version ${PKG_VERS}" \
             -D "release ${PKG_REV}" \
             packaging/rpm/SPECS/hcu-container-toolkit.spec
    # 检查并移动RPM包，添加错误提示
    mv RPMS/"$arch"/*.rpm dist || { echo "Error: 未找到生成的RPM包" >&2; exit 1; }

fi


if [ "$PACKAGE_TYPE" = "deb" ]; then
   ARCH="$(dpkg-architecture -qDEB_HOST_ARCH)"

   if [ -d "${PKG_NAME}" ];then
      rm -rf "${PKG_NAME}"
   fi

   if [ ! -d dist ];then
      mkdir dist
   fi

   mkdir -p "${PKG_NAME}/DEBIAN"
   mkdir -p "${PKG_NAME}/usr/bin"

   cp  hcu-container-runtime "${PKG_NAME}/usr/bin/"
   cp  hcu-cdi-hook "${PKG_NAME}/usr/bin/"
   cp  hcu-ctk "${PKG_NAME}/usr/bin/"
   cp  hcu-docker "${PKG_NAME}/usr/bin/"

   chmod 755 "${PKG_NAME}"/usr/bin/*

   cp packaging/debian/hcu-container-toolkit.postinst "${PKG_NAME}/DEBIAN/postinst"
   chmod 755 "${PKG_NAME}/DEBIAN/postinst"

   cp packaging/debian/hcu-container-toolkit.postrm "${PKG_NAME}/DEBIAN/postrm"
   chmod 755 "${PKG_NAME}/DEBIAN/postrm"

   cat > "${PKG_NAME}/DEBIAN/control" << EOF
Package: ${PKG_NAME}
Version: ${PACKAGE_VERSION}-${PACKAGE_REVISION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: Your Name <you@example.com>
Description: hcu runtime and related utilities
 This package contains hcu-container-runtime, hcu-ctk, hcu-docker, hcu-cdi-hook.
EOF



   dpkg-deb --build "${PKG_NAME}" "dist/${PKG_NAME}_${PACKAGE_VERSION}-${PACKAGE_REVISION}_${ARCH}.deb"
   rm -rf "${PKG_NAME}"
   rm -rf hcu-container-runtime hcu-ctk hcu-cdi-hook hcu-docker
fi
