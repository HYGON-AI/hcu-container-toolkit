/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package hcuTracker

import (
	"encoding/json"
	"fmt"
	"hcu-container-toolkit/internal/hyhcu"
	"os"
	"os/signal"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofrs/flock"
)

type accessibility int

const (
	SHARED_ACCESS accessibility = iota
	EXCLUSIVE_ACCESS
)

// Interface for HCU Tracker package

type Interface interface {
	// Initialize HCU Tracker
	Init() error
	// Enable HCU Tracker
	Enable() error
	// Disable HCU Tracker
	Disable() error
	// Reset HCU Tracker
	Reset() error
	// Show HCUs  status
	ShowStatus() error

	// Make specified HCUs exclusive such that they can be used
	// by at most one container at any instance
	MakeHCUsExclusive(hcus string) error

	// Make specified HCUs shared such that they can be used
	// by any number of containers at any instance
	MakeHCUsShared(hcus string) error

	// Reserve HCUs for a container
	ReserveHCUs(hcus string, containerId string) ([]int, error)

	// Release all HCUs linked to a container
	ReleaseHCUs(containerId string) error
}

type hcu_status_t struct {
	// UUID of HCU
	UUID string `json:"uuid"`

	// Partition Type of the HCU
	PartitionType string `json:"partitionType"`

	// HCU accessibility
	Accessibility accessibility `json:"accessibility"`

	// Container Ids of the containers to which the HCU is assigned
	ContainerIds []string `json:"containerIds"`
}

type hcu_tracker_data_t struct {
	//Status of HCU Tracker
	Enabled bool `json:"enabled"`

	//Status of all HCUs
	HCUsStatus map[int]hcu_status_t `json:"hcusStatus"`

	// Info of all HCUs
	HCUsInfo map[int]hyhcu.DeviceInfo `json:"hcusInfo"`
}

// isHCUTrackerInitializedType is the type for functions
// that return if HCU Tracker is initialized
type isHCUTrackerInitializedType func() (bool, error)

// initializeHCUTrackerType is the type for functions that
// initialize HCU Tracker
type initializeHCUTrackerType func() error

// parseHCUsListType is the type for functions that parse
// HCU list strings and returns the valid and invalid HCU Ids
type parseHCUsListType func(string) ([]int, []string, []string, error)

// readHCUTrackerFileType is the type for functions that
// read the HCU Tracker file and return the HCUs status
type readHCUTrackerFileType func() (hcu_tracker_data_t, error)

// writeHCUTrackerFileType is the type for functions that
// write the HCUs status to HCU Tracker file
type writeHCUTrackerFileType func(hcu_tracker_data_t) error

// validateHCUsInfoType is the type for functions that
// validate the HCUs info
type validateHCUsInfoType func(map[int]hyhcu.DeviceInfo) (bool, error)

type hcu_tracker_t struct {
	// path to HCU Tracker lock file
	hcuTrackerLockFile string

	// function to check if HCU Tracker is initialized
	isHCUTrackerInitialized isHCUTrackerInitializedType

	// function to initialize HCU Tracker
	initializeHCUTracker initializeHCUTrackerType

	// function to parse HCU list strings
	parseHCUsList parseHCUsListType

	// function to read HCU Tracker file
	readHCUTrackerFile readHCUTrackerFileType

	// function to write HCU Tracker file
	writeHCUTrackerFile writeHCUTrackerFileType

	// function to validate HCUs info
	validateHCUsInfo validateHCUsInfoType
}

const (
	hcuTrackerFile     = "/var/log/hcu-tracker.json"
	hcuTrackerLockFile = "/var/log/hcu-tracker.lock"
)

func setupSignalHandler(lock *flock.Flock) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-c
		fmt.Printf("Received signal: %v. Cleaning up...\n", sig)
		if lock != nil {
			_ = lock.Unlock()
		}
		os.Exit(1)
	}()
}

func acquireLock(lockFile string) (*flock.Flock, error) {
	lock := flock.New(lockFile)

	timeout := time.After(10 * time.Second)
	tick := time.Tick(100 * time.Millisecond)

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("Acquiring lock timeout exceeded")
		case <-tick:
			locker, err := lock.TryLock()
			if err != nil {
				return nil, fmt.Errorf("Failed to acquire lock, Error: %v", err)
			}
			if locker {
				return lock, nil
			}
		}
	}
}

func parseHCUsList(hcus string) ([]int, []string, []string, error) {
	// isHexString checks if a string contains only hexadecimal characters
	isHexString := func(s string) bool {
		if len(s) == 0 {
			return false
		}
		for _, c := range s {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
		return true
	}

	validHCUs := []int{}
	invalidHCUs := []string{}
	invalidHCUsRange := []string{}

	hcusInfo, err := hyhcu.GetHYHCUs()
	if err != nil {
		return []int{}, []string{}, []string{}, fmt.Errorf("Failed to get HCU info, Error: %v", err)
	}

	if hcus == "all" || hcus == "All" || hcus == "ALL" {
		for i := 0; i < len(hcusInfo); i++ {
			validHCUs = append(validHCUs, i)
		}
		return validHCUs, []string{}, []string{}, nil
	}

	uuidToHCUIdMap, err := hyhcu.GetUniqueIdToDeviceIndexMap()

	if err != nil {
		fmt.Printf("Failed to get UUID to HCU Id mappings: %v", err)
		uuidToHCUIdMap = make(map[string][]int)
	}

	for _, c := range strings.Split(hcus, ",") {
		if strings.HasPrefix(c, "0x") || strings.HasPrefix(c, "0X") ||
			(len(c) > 8 && isHexString(c)) {
			uuid := strings.ToLower(c)
			if !strings.HasPrefix(uuid, "0x") {
				uuid = "0x" + uuid
			}
			if hcuIds, exists := uuidToHCUIdMap[uuid]; exists {
				validHCUs = append(validHCUs, hcuIds...)
			} else {
				uuid = strings.TrimPrefix(uuid, "0x")
				if hcuIds, exists := uuidToHCUIdMap[uuid]; exists {
					validHCUs = append(validHCUs, hcuIds...)
				} else {
					invalidHCUs = append(invalidHCUs, c)
				}
			}
		} else if strings.Contains(c, "-") {
			devsRange := strings.SplitN(c, "-", 2)
			start, err0 := strconv.Atoi(devsRange[0])
			end, err1 := strconv.Atoi(devsRange[1])
			if err0 != nil || err1 != nil ||
				start < 0 || end < 0 || start > end {
				invalidHCUsRange = append(invalidHCUsRange, c)
			} else {
				for i := start; i <= end; i++ {
					if i < len(hcusInfo) {
						validHCUs = append(validHCUs, i)
					} else {
						invalidHCUs = append(invalidHCUs, strconv.Itoa(i))
					}
				}
			}
		} else {
			i, err := strconv.Atoi(c)
			if err == nil {
				if i >= 0 && i < len(hcusInfo) {
					validHCUs = append(validHCUs, i)
				} else {
					invalidHCUs = append(invalidHCUs, c)
				}
			} else {
				invalidHCUs = append(invalidHCUs, c)
			}
		}
	}

	sort.Ints(validHCUs)

	return validHCUs, invalidHCUs, invalidHCUsRange, nil
}

func isHCUTrackerInitialized() (bool, error) {
	hcuTrackerInitialized := false
	_, err := os.Stat(hcuTrackerFile)
	if err == nil {
		hcuTrackerInitialized = true
	} else {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("Error checking file %v, Error:%v", hcuTrackerFile, err)
		}
	}
	return hcuTrackerInitialized, nil
}

func readHCUTrackerFile() (hcu_tracker_data_t, error) {
	file, err := os.Open(hcuTrackerFile)
	if err != nil {
		return hcu_tracker_data_t{HCUsStatus: make(map[int]hcu_status_t), HCUsInfo: make(map[int]hyhcu.DeviceInfo)},
			fmt.Errorf("Error opening file, Error: %v", err)
	}
	defer file.Close()

	var hcuTrackerData hcu_tracker_data_t
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&hcuTrackerData); err != nil {
		return hcu_tracker_data_t{HCUsStatus: make(map[int]hcu_status_t), HCUsInfo: make(map[int]hyhcu.DeviceInfo)},
			fmt.Errorf("Failed to decode JSON, Error: %v", err)
	}
	return hcuTrackerData, nil
}

func writeHCUTrackerFile(hcuTrackerData hcu_tracker_data_t) error {
	tempPath := hcuTrackerFile + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("Error creating temp file, Error: %v", err)
	}

	encoder := json.NewEncoder(tempFile)
	if err := encoder.Encode(hcuTrackerData); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return fmt.Errorf("Error encoding JSON to temp file, Error: %v", err)
	}

	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return fmt.Errorf("Error syncing temp file: %v", err)
	}

	tempFile.Close()

	if err := os.Rename(tempPath, hcuTrackerFile); err != nil {
		return fmt.Errorf("Error renaming temp file: %v", err)
	}
	return nil
}

func initializeHCUTracker() error {
	hcusInfo, err := hyhcu.GetHYHCUs()
	if err != nil {
		return fmt.Errorf("Failed to get HY HCUs info, Error: %v", err)
	}

	uuidToHCUIdMap, err := hyhcu.GetUniqueIdToDeviceIndexMap()
	if err != nil {
		uuidToHCUIdMap = make(map[string][]int) // Continue with empty map
	}
	hcuIdToUUIDMap := make(map[int]string)
	for uuid, hcuIds := range uuidToHCUIdMap {
		if strings.HasPrefix(uuid, "0x") || strings.HasPrefix(uuid, "0X") {
			uuid = uuid[2:]
		}
		uuid = "0x" + strings.ToUpper(uuid)
		for _, hcuId := range hcuIds {
			hcuIdToUUIDMap[hcuId] = uuid
		}
	}

	hcuTrackerData := hcu_tracker_data_t{Enabled: false, HCUsStatus: make(map[int]hcu_status_t), HCUsInfo: make(map[int]hyhcu.DeviceInfo)}
	for hcuId, hcuInfo := range hcusInfo {
		hcuTrackerData.HCUsInfo[hcuId] = hcuInfo
		hcuTrackerData.HCUsStatus[hcuId] = hcu_status_t{
			UUID:          hcuIdToUUIDMap[hcuId],
			PartitionType: hcusInfo[hcuId].PartitionType,
			Accessibility: SHARED_ACCESS,
			ContainerIds:  []string{},
		}
	}

	return writeHCUTrackerFile(hcuTrackerData)
}

func validateHCUsInfo(savedHCUsInfo map[int]hyhcu.DeviceInfo) (bool, error) {
	tempHCUsInfo, err := hyhcu.GetHYHCUs()
	if err != nil {
		return false, fmt.Errorf("Failed to get HY HCUs info, Error: %v", err)
	}
	currentHCUsInfo := make(map[int]hyhcu.DeviceInfo)
	for hcuId, hcuInfo := range tempHCUsInfo {
		currentHCUsInfo[hcuId] = hcuInfo
	}

	equal := reflect.DeepEqual(savedHCUsInfo, currentHCUsInfo)
	if equal != true {
		fmt.Printf("HCUs info is invalid. Please reset HCU Tracker.\n")
		return false, nil
	}

	return true, nil
}

func (hcuTracker *hcu_tracker_t) Init() (err error) {
	lock, err := acquireLock(hcuTracker.hcuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("Init lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic in Init: %v", r)
		}
	}()

	err = hcuTracker.initializeHCUTracker()
	if err != nil {
		return fmt.Errorf("Failed to initialize HCU Tracker, Error: %v", err)
	}

	return nil
}

func (hcuTracker *hcu_tracker_t) Enable() (err error) {
	lock, err := acquireLock(hcuTracker.hcuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("Enable lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic in Enable: %v", r)
		}
	}()

	hcuTrackerInitialized, err := hcuTracker.isHCUTrackerInitialized()
	if err != nil {
		return fmt.Errorf("Failed to check if HCU Tracker is initialized, Error: %v\n", err)
	}

	if !hcuTrackerInitialized {
		err := hcuTracker.initializeHCUTracker()
		if err != nil {
			return err
		}
	}

	hcusTrackerData, err := hcuTracker.readHCUTrackerFile()
	if err != nil {
		fmt.Printf("Failed to show HCU Tracker status, Error: %v\n", err)
		return err
	}

	if hcusTrackerData.Enabled {
		fmt.Printf("HCU Tracker is already enabled\n")
		return nil
	}

	err = hcuTracker.initializeHCUTracker()
	if err != nil {
		fmt.Printf("Failed to enable HCU Tracker, Error: %v\n", err)
		return err
	}

	hcusTrackerData, err = hcuTracker.readHCUTrackerFile()
	if err != nil {
		fmt.Printf("Failed to enable HCU Tracker, Error: %v\n", err)
		return err
	}

	hcusTrackerData.Enabled = true

	err = hcuTracker.writeHCUTrackerFile(hcusTrackerData)
	if err != nil {
		fmt.Printf("Failed to enable HCU Tracker, Error: %v\n", err)
		return err
	}

	fmt.Printf("HCU Tracker has been enabled\n")
	return nil
}

func (hcuTracker *hcu_tracker_t) Disable() (err error) {
	lock, err := acquireLock(hcuTracker.hcuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("Disable lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic in Disable: %v", r)
		}
	}()

	hcuTrackerInitialized, err := hcuTracker.isHCUTrackerInitialized()
	if err != nil {
		fmt.Printf("Failed to check if HCU Tracker is initialized, Error:%v\n", err)
		return err
	}

	if !hcuTrackerInitialized {
		err := hcuTracker.initializeHCUTracker()
		if err != nil {
			fmt.Printf("Failed to disable HCU Tracker, Error: %v\n", err)
			return err
		}
	} else {
		hcusTrackerData, err := hcuTracker.readHCUTrackerFile()
		if err != nil {
			fmt.Printf("Failed to disable HCU Tracker, Error: %v\n", err)
			return err
		}

		hcusTrackerData.Enabled = false

		err = hcuTracker.writeHCUTrackerFile(hcusTrackerData)
		if err != nil {
			fmt.Printf("Failed to disable HCU Tracker, Error: %v\n", err)
			return err
		}
	}

	fmt.Printf("HCU Tracker has been disabled\n")
	return nil
}

func (hcuTracker *hcu_tracker_t) Reset() (err error) {
	lock, err := acquireLock(hcuTracker.hcuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("Reset lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic in Reset: %v", r)
		}
	}()

	hcuTrackerInitialized, err := hcuTracker.isHCUTrackerInitialized()
	if err != nil {
		fmt.Printf("Failed to check if HCU Tracker is initialized, Error:%v\n", err)
		return err
	}

	hcuTrackerEnabled := false

	if !hcuTrackerInitialized {
		err := hcuTracker.initializeHCUTracker()
		if err != nil {
			fmt.Printf("Failed to reset HCU Tracker, Error: %v\n", err)
			return err
		}
	} else {
		hcusTrackerData, err := hcuTracker.readHCUTrackerFile()
		if err != nil {
			fmt.Printf("Failed to reset HCU Tracker, Error: %v\n", err)
			return err
		}

		hcuTrackerEnabled = hcusTrackerData.Enabled

		err = hcuTracker.initializeHCUTracker()
		if err != nil {
			fmt.Printf("Failed to reset HCU Tracker, Error: %v\n", err)
			return err
		}

		hcusTrackerData, err = hcuTracker.readHCUTrackerFile()
		if err != nil {
			fmt.Printf("Failed to reset HCU Tracker, Error: %v\n", err)
			return err
		}

		if hcuTrackerEnabled == true {
			hcusTrackerData.Enabled = true
			err = hcuTracker.writeHCUTrackerFile(hcusTrackerData)
			if err != nil {
				fmt.Printf("Failed to reset HCU Tracker, Error: %v\n", err)
				return err
			}
		}
	}

	fmt.Printf("HCU Tracker has been reset\n")
	if hcuTrackerEnabled {
		fmt.Printf("Since HCU Tracker was enabled, it is recommended to stop and restart running containers to get the most accurate HCU Tracker status\n")
	}
	return nil
}

func (hcuTracker *hcu_tracker_t) ShowStatus() (err error) {
	lock, err := acquireLock(hcuTracker.hcuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("ShowStatus lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic in ShowStatus: %v", r)
		}
	}()

	hcuTrackerInitialized, err := hcuTracker.isHCUTrackerInitialized()
	if err != nil {
		fmt.Printf("Failed to check if HCU Tracker is initialized, Error:%v\n", err)
		return err
	}

	if !hcuTrackerInitialized {
		err := hcuTracker.initializeHCUTracker()
		if err != nil {
			return err
		}
	}

	hcusTrackerData, err := hcuTracker.readHCUTrackerFile()
	if err != nil {
		fmt.Printf("Failed to show HCU Tracker status, Error: %v\n", err)
		return err
	}

	if hcusTrackerData.Enabled == false {
		fmt.Printf("HCU Tracker is disabled\n")
		return nil
	}

	result, err := hcuTracker.validateHCUsInfo(hcusTrackerData.HCUsInfo)
	if err != nil || result != true {
		return err
	}

	fmt.Println(strings.Repeat("-", 120))
	fmt.Printf("%-10s%-25s%-20s%-65s\n", "HCU Id", "UUID", "Accessibility", "Container Ids")
	fmt.Println(strings.Repeat("-", 120))
	for hcuId := 0; hcuId < len(hcusTrackerData.HCUsStatus); hcuId++ {
		var accessibility string
		switch hcusTrackerData.HCUsStatus[hcuId].Accessibility {
		case SHARED_ACCESS:
			accessibility = "Shared"
		case EXCLUSIVE_ACCESS:
			accessibility = "Exclusive"
		default:
			fmt.Printf("Invalid accessibility value %v\n", hcusTrackerData.HCUsStatus[hcuId].Accessibility)
			break
		}

		if len(hcusTrackerData.HCUsStatus[hcuId].ContainerIds) > 0 {
			for idx, id := range hcusTrackerData.HCUsStatus[hcuId].ContainerIds {
				if idx == 0 {
					fmt.Printf("%-10v%-25s%-20v%-65v\n", hcuId, hcusTrackerData.HCUsStatus[hcuId].UUID, accessibility, id)
				} else {
					fmt.Printf("%-10v%-25v%-20v%-65v\n", "", "", "", id)
				}
			}
		} else {
			fmt.Printf("%-10v%-25v%-20v%-65v\n", hcuId, hcusTrackerData.HCUsStatus[hcuId].UUID, accessibility, "None")
		}
	}

	return nil
}

func (hcuTracker *hcu_tracker_t) MakeHCUsExclusive(hcus string) (err error) {
	lock, err := acquireLock(hcuTracker.hcuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("MakeHCUsExclusive lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic in MakeHCUsExclusive: %v", r)
		}
	}()

	hcuTrackerInitialized, err := hcuTracker.isHCUTrackerInitialized()
	if err != nil {
		fmt.Printf("Failed to check if HCU Tracker is initialized, Error:%v\n", err)
		return err
	}

	if !hcuTrackerInitialized {
		err := hcuTracker.initializeHCUTracker()
		if err != nil {
			return err
		}
	}

	hcusTrackerData, err := hcuTracker.readHCUTrackerFile()
	if err != nil {
		fmt.Printf("Failed to make HCUs exclusive, Error: %v\n", err)
		return err
	}

	if hcusTrackerData.Enabled == false {
		fmt.Printf("HCU Tracker is disabled\n")
		return nil
	}

	result, err := hcuTracker.validateHCUsInfo(hcusTrackerData.HCUsInfo)
	if err != nil || result != true {
		return err
	}

	validHCUs, invalidHCUs, invalidHCUsRange, err := hcuTracker.parseHCUsList(hcus)
	if err != nil {
		fmt.Printf("Failed to parse HCUs list, Error: %v\n", err)
		return err
	}

	hcusMadeExclusive := []int{}
	hcusNotMadeExclusive := []int{}

	for _, hcuId := range validHCUs {
		if len(hcusTrackerData.HCUsStatus[hcuId].ContainerIds) < 2 {
			hcusTrackerData.HCUsStatus[hcuId] = hcu_status_t{
				UUID:          hcusTrackerData.HCUsStatus[hcuId].UUID,
				PartitionType: hcusTrackerData.HCUsStatus[hcuId].PartitionType,
				Accessibility: EXCLUSIVE_ACCESS,
				ContainerIds:  hcusTrackerData.HCUsStatus[hcuId].ContainerIds,
			}
			hcusMadeExclusive = append(hcusMadeExclusive, hcuId)
		} else {
			hcusNotMadeExclusive = append(hcusNotMadeExclusive, hcuId)
		}
	}

	err = hcuTracker.writeHCUTrackerFile(hcusTrackerData)
	if err != nil {
		fmt.Printf("Failed to make HCUs exclusive, Error: %v\n", err)
		return err
	}

	if len(hcusMadeExclusive) > 0 {
		fmt.Printf("HCUs %v have been made exclusive\n", hcusMadeExclusive)
	}
	if len(hcusNotMadeExclusive) > 0 {
		fmt.Printf("HCUs %v have not been made exclusive because more than one container is currently using it\n", hcusNotMadeExclusive)
	}
	if len(invalidHCUsRange) > 0 {
		fmt.Printf("Ignoring %v HCUs Ranges as they are invalid\n", invalidHCUsRange)
	}
	if len(invalidHCUs) > 0 {
		fmt.Printf("Ignoring %v HCUs as they are invalid\n", invalidHCUs)
	}

	return nil
}

func (hcuTracker *hcu_tracker_t) MakeHCUsShared(hcus string) (err error) {
	lock, err := acquireLock(hcuTracker.hcuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("MakeHCUsShared lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic in MakeHCUsShared: %v", r)
		}
	}()

	hcuTrackerInitialized, err := hcuTracker.isHCUTrackerInitialized()

	if err != nil {
		fmt.Printf("Failed to check if HCU Tracker is initialized, Error:%v\n", err)
		return err
	}

	if !hcuTrackerInitialized {
		err = hcuTracker.initializeHCUTracker()
		if err != nil {
			return err
		}
	}

	hcusTrackerData, err := hcuTracker.readHCUTrackerFile()
	if err != nil {
		fmt.Printf("Failed to make HCUs %v shared, Error: %v\n", hcus, err)
		return err
	}

	if hcusTrackerData.Enabled == false {
		fmt.Printf("HCU Tracker is disabled\n")
		return nil
	}

	result, err := hcuTracker.validateHCUsInfo(hcusTrackerData.HCUsInfo)
	if err != nil || result != true {
		return err
	}

	validHCUs, invalidHCUs, invalidHCUsRange, err := hcuTracker.parseHCUsList(hcus)
	if err != nil {
		fmt.Printf("Failed to parse HCUs list %v, Error: %v\n", hcus, err)
		return err
	}

	for _, hcuId := range validHCUs {
		hcusTrackerData.HCUsStatus[hcuId] = hcu_status_t{
			UUID:          hcusTrackerData.HCUsStatus[hcuId].UUID,
			PartitionType: hcusTrackerData.HCUsStatus[hcuId].PartitionType,
			Accessibility: SHARED_ACCESS,
			ContainerIds:  hcusTrackerData.HCUsStatus[hcuId].ContainerIds,
		}
	}

	err = hcuTracker.writeHCUTrackerFile(hcusTrackerData)
	if err != nil {
		fmt.Printf("Failed to make HCUs shared, Error: %v\n", err)
		return err
	}

	if len(validHCUs) > 0 {
		fmt.Printf("HCUs %v have been made shared\n", validHCUs)
	}
	if len(invalidHCUsRange) > 0 {
		fmt.Printf("Ignoring %v HCUs Ranges as they are invalid\n", invalidHCUsRange)
	}
	if len(invalidHCUs) > 0 {
		fmt.Printf("Ignoring %v HCUs as they are invalid\n", invalidHCUs)
	}

	return nil
}

func (hcuTracker *hcu_tracker_t) ReserveHCUs(hcus string, containerId string) (allocatedHCUs []int, err error) {
	lock, err := acquireLock(hcuTracker.hcuTrackerLockFile)
	if err != nil {
		return nil, fmt.Errorf("ReserveHCUs lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic in ReserveHCUs: %v", r)
			allocatedHCUs = []int{}
		}
	}()

	hcuTrackerInitialized, err := hcuTracker.isHCUTrackerInitialized()
	if err != nil {
		fmt.Printf("Failed to check if HCU Tracker is initialized, Error:%v\n", err)
		return []int{}, err
	}

	if !hcuTrackerInitialized {
		err = hcuTracker.initializeHCUTracker()
		if err != nil {
			return []int{}, err
		}
	}

	hcusTrackerData, err := hcuTracker.readHCUTrackerFile()
	if err != nil {
		fmt.Printf("Failed to reserve HCUs %v, Error:%v\n", hcus, err)
		return []int{}, err
	}

	validHCUs, invalidHCUs, invalidHCUsRange, err := hcuTracker.parseHCUsList(hcus)
	if err != nil {
		fmt.Printf("Failed to parse HCUs list %v, Error: %v\n", hcus, err)
		return []int{}, err
	}
	if len(invalidHCUsRange) > 0 {
		fmt.Printf("Ignoring %v HCUs Ranges as they are invalid\n", invalidHCUsRange)
	}

	if len(invalidHCUs) > 0 {
		fmt.Printf("Ignoring %v HCUs as they are invalid\n", invalidHCUs)
	}

	if hcusTrackerData.Enabled == false {
		return validHCUs, nil
	}

	result, err := hcuTracker.validateHCUsInfo(hcusTrackerData.HCUsInfo)
	if err != nil || result != true {
		return []int{}, fmt.Errorf("HCUs info is invalid, Please reset HCU Tracker.\n")
	}

	var unavailableHCUs []int
	for _, hcuId := range validHCUs {
		if hcusTrackerData.HCUsStatus[hcuId].Accessibility == SHARED_ACCESS ||
			(hcusTrackerData.HCUsStatus[hcuId].Accessibility == EXCLUSIVE_ACCESS &&
				len(hcusTrackerData.HCUsStatus[hcuId].ContainerIds) == 0) {
			hcusTrackerData.HCUsStatus[hcuId] = hcu_status_t{
				UUID:          hcusTrackerData.HCUsStatus[hcuId].UUID,
				PartitionType: hcusTrackerData.HCUsStatus[hcuId].PartitionType,
				Accessibility: hcusTrackerData.HCUsStatus[hcuId].Accessibility,
				ContainerIds:  append(hcusTrackerData.HCUsStatus[hcuId].ContainerIds, containerId),
			}
			allocatedHCUs = append(allocatedHCUs, hcuId)
		} else {
			unavailableHCUs = append(unavailableHCUs, hcuId)
		}
	}
	err = hcuTracker.writeHCUTrackerFile(hcusTrackerData)
	if err != nil {
		fmt.Printf("Failed to reserve HCUs %v, Error:%v\n", validHCUs, err)
		return []int{}, err
	}

	if len(allocatedHCUs) > 0 {
		fmt.Printf("HCUs %v allocated\n", allocatedHCUs)
	}
	if len(unavailableHCUs) > 0 {
		fmt.Printf("HCUs %v are exclusive and already in use\n", unavailableHCUs)
		return []int{}, fmt.Errorf("HCUs %v are exclusive and already in use\n", unavailableHCUs)
	}
	return allocatedHCUs, nil
}

func (hcuTracker *hcu_tracker_t) ReleaseHCUs(containerId string) (err error) {
	removeContainerId := func(containerId string, containerIds []string) ([]string, bool) {
		for idx, id := range containerIds {
			if id == containerId {
				return append(containerIds[:idx], containerIds[idx+1:]...), true
			}
		}
		return containerIds, false
	}

	lock, err := acquireLock(hcuTracker.hcuTrackerLockFile)
	if err != nil {
		return fmt.Errorf("ReleaseHCUs lock failed: %v", err)
	}

	defer func() {
		if lock != nil {
			_ = lock.Unlock()
		}
	}()
	setupSignalHandler(lock)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic in ReleaseHCUs: %v", r)
		}
	}()

	hcuTrackerInitialized, err := hcuTracker.isHCUTrackerInitialized()
	if err != nil {
		fmt.Printf("Failed to check if HCU Tracker is initialized, Error:%v\n", err)
		return err
	}

	if hcuTrackerInitialized {
		hcusTrackerData, err := hcuTracker.readHCUTrackerFile()
		if err != nil {
			fmt.Printf("Failed to release HCUs used by container %v, Error: %v\n", containerId, err)
			return err
		}
		var releasedHCUs []int
		for hcuId, _ := range hcusTrackerData.HCUsStatus {
			containerIds, released := removeContainerId(containerId, hcusTrackerData.HCUsStatus[hcuId].ContainerIds)
			if released {
				hcusTrackerData.HCUsStatus[hcuId] = hcu_status_t{
					UUID:          hcusTrackerData.HCUsStatus[hcuId].UUID,
					PartitionType: hcusTrackerData.HCUsStatus[hcuId].PartitionType,
					Accessibility: hcusTrackerData.HCUsStatus[hcuId].Accessibility,
					ContainerIds:  containerIds,
				}
				releasedHCUs = append(releasedHCUs, hcuId)
			}
		}

		err = hcuTracker.writeHCUTrackerFile(hcusTrackerData)
		if err != nil {
			fmt.Printf("Failed to release HCUs used by container %v, Error: %v\n", containerId, err)
			return err
		}

		fmt.Printf("Released HCUs %v used by container %v\n", releasedHCUs, containerId)
	}

	return nil

}

func New() (Interface, error) {
	hcuTracker := &hcu_tracker_t{
		hcuTrackerLockFile:      hcuTrackerLockFile,
		isHCUTrackerInitialized: isHCUTrackerInitialized,
		initializeHCUTracker:    initializeHCUTracker,
		parseHCUsList:           parseHCUsList,
		readHCUTrackerFile:      readHCUTrackerFile,
		writeHCUTrackerFile:     writeHCUTrackerFile,
		validateHCUsInfo:        validateHCUsInfo,
	}
	return hcuTracker, nil
}
