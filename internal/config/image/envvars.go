/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package image

const (
	EnvVarHCUVisibleDevices    = "HCU_VISIBLE_DEVICES"
	EnvVarNvidiaVisibleDevices = "NVIDIA_VISIBLE_DEVICES"
	EnvVarROCmVersion          = "ROCM_VERSION"
	EnvVarVHCUVisibleDevices   = "VHCU_VISIBLE_DEVICES"
	EnvVarMIGVisibleDevices    = "MIG_VISIBLE_DEVICES"
	EnvVarDevicePluginMode     = "RESOURCE_REGISTER_STRATEGY"

	// Dynamic vHCU split request variables. Setting either Cores or Mem
	// triggers a dynamic vHCU allocation, while Mem is mandatory because a
	// cores-only request cannot express the memory to carve out; see
	// internal/vhcu.
	EnvVarVHCUCores  = "HCU_VHCU_CORES"
	EnvVarVHCUMem    = "HCU_VHCU_MEM"
	EnvVarVHCUDevice = "HCU_VHCU_DEVICE"
)
