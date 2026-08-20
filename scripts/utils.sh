#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2026 Hygon Information Technology Co., Ltd.
#

# This list represents the distribution-architecture pairs that are actually published
# to the relevant repositories. This targets forwarded to the build-all-components script
# can be overridden by specifying command line arguments.
# shellcheck disable=SC2034
all=(
    ubuntu18.04-amd64
    ubuntu20.04-amd64
    ubuntu22.04-amd64
    centos7-x86_64
    centos8-x86_64
    rocky8-x86_64
    rocky9-x86_64
    kylin10-x86_64
)

function get_base_image() {
    local baseimage
    case ${1} in
    ubuntu18.04-amd64)
        baseimage=${2}/ubuntu:18.04
        ;;
    ubuntu20.04-amd64)
        baseimage=${2}/ubuntu:20.04
        ;;
    ubuntu22.04-amd64)
        baseimage=${2}/ubuntu:22.04
        ;;
    centos7-x86_64)
        baseimage=${2}/centos:7
        ;;
    centos8-x86_64)
        baseimage=${2}/centos:8.4.2105
        ;;
    rocky8-x86_64)
        baseimage=${2}/rockylinux:8.6
        ;;
    rocky9-x86_64)
        baseimage=${2}/rockylinux:9.2
        ;;
    kylin10-x86_64)
        baseimage=${2}/kylin:v10-sp2
        ;;
    esac
    echo "${baseimage}"
}
