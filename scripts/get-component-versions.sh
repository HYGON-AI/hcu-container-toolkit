#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2026 Hygon Information Technology Co., Ltd.
#

function assert_usage() {
    exit 1
}

set -e

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/../scripts && pwd )"
PROJECT_ROOT="$( cd "${SCRIPTS_DIR}/.." && pwd )"

HCU_CONTAINER_TOOLKIT_ROOT=${PROJECT_ROOT}

versions_makefile="${HCU_CONTAINER_TOOLKIT_ROOT}/versions.mk"

hcu_container_toolkit_version=$(grep -m 1 "^LIB_VERSION := " "${versions_makefile}" | sed -e 's/LIB_VERSION :=[[:space:]]\(.*\)[[:space:]]*/\1/')
hcu_container_toolkit_tag=$(grep -m 1 "^LIB_TAG .= " "${versions_makefile}" | sed -e 's/LIB_TAG .=[[:space:]]\(.*\)[[:space:]]*/\1/')
hcu_container_toolkit_version_tag="${hcu_container_toolkit_version}${hcu_container_toolkit_tag}:+~${hcu_container_toolkit_tag}"

echo "HCU_CONTAINER_TOOLKIT_VERSION=${hcu_container_toolkit_version}"
echo "HCU_CONTAINER_TOOLKIT_TAG=${hcu_container_toolkit_tag}"
echo "HCU_CONTAINER_TOOLKIT_PACKAGE_VERSION=${hcu_container_toolkit_version_tag//\~/-}"
