/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package main

import (
	"hcu-container-toolkit/internal/hcu-tracker"
	"hcu-container-toolkit/internal/runtime"
	"os"
	//"log"
)

func main() {
	//f, _ := os.OpenFile("/var/log/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	//log.SetOutput(f)
	r := runtime.New()
	//log.Println("Log written to file")
	err := r.Run(os.Args)
	if err != nil {
		//	log.Println("failed to run runtime")
		hcuTracker, err := hcuTracker.New()
		if err != nil {
			//		log.Println("failed to new hcutracker")
			os.Exit(1)
		}
		hcuTracker.ReleaseHCUs(os.Args[len(os.Args)-1])
		//	log.Println("releasehcu")
		os.Exit(1)
	}
}
