#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2026 Hygon Information Technology Co., Ltd.
#

set -e

export DOCKER_IMAGE_REGISTRY=10.65.42.71:9092

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )"/../scripts && pwd )"

# shellcheck source=scripts/utils.sh
source "${SCRIPTS_DIR}"/utils.sh

if [[ $# -gt 0 ]]; then
    targets=("$@")
else
    targets=("${all[@]}")
fi

eval "$("${SCRIPTS_DIR}"/get-component-versions.sh)"

export HCU_CONTAINER_TOOLKIT_VERSION
export HCU_CONTAINER_TOOLKIT_TAG

for target in "${targets[@]}"; do

    if [[ -z "${DOCKER_IMAGE_REGISTRY}" ]]; then
        image_args=
    else
        baseimage=$(get_base_image "${target}" "${DOCKER_IMAGE_REGISTRY}")
        image_args="BASEIMAGE=${baseimage}"
    fi
    "${SCRIPTS_DIR}/build-all-components.sh" "${target}" "${image_args}"
done
