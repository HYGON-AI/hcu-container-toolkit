#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2026 Hygon Information Technology Co., Ltd.
#

# This script is used to build the packages for the components of the C-3000 Container Stack.

set -e

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/../scripts && pwd )"
PROJECT_ROOT="$( cd "${SCRIPTS_DIR}"/.. && pwd )"

function assert_usage() {
    echo "Missing argument $1"
    echo "$(basename "${BASH_SOURCE[0]}") TARGET"
    exit 1
}

if [[ $# -le 1 ]]; then
    assert_usage "TARGET"
fi

TARGET=$1
BASEIMAGE=$2

# shellcheck source=scripts/utils.sh
source "${SCRIPTS_DIR}"/utils.sh

: "${DIST_DIR:=${PROJECT_ROOT}/dist}"
export DIST_DIR

echo "Building ${TARGET} for all packages to ${DIST_DIR}"

: "${HCU_CONTAINER_TOOLKIT_ROOT:=${PROJECT_ROOT}}"

"${SCRIPTS_DIR}/get-component-versions.sh"

if [[ -z "${HCU_CONTAINER_TOOLKIT_VERSION}" ]]; then
eval "$("${SCRIPTS_DIR}"/get-component-versions.sh)"
fi

make -C "${HCU_CONTAINER_TOOLKIT_ROOT}" "${TARGET}" "${BASEIMAGE}"