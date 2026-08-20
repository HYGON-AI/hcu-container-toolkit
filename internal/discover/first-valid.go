/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

import "errors"

type firstOf []Discover

// FirstValid returns a discoverer that returns the first non-error result from a list of discoverers.
func FirstValid(discoverers ...Discover) Discover {
	var f firstOf
	for _, d := range discoverers {
		if d == nil {
			continue
		}
		f = append(f, d)
	}
	return f
}

func (f firstOf) Devices() ([]Device, error) {
	var errs error
	for _, d := range f {
		devices, err := d.Devices()
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		return devices, nil
	}
	return nil, errs
}

func (f firstOf) Hooks() ([]Hook, error) {
	var errs error
	for _, d := range f {
		hooks, err := d.Hooks()
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		return hooks, nil
	}
	return nil, errs
}

func (f firstOf) Mounts() ([]Mount, error) {
	var errs error
	for _, d := range f {
		mounts, err := d.Mounts()
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		return mounts, nil
	}
	return nil, nil
}

func (f firstOf) AdditionalGIDs() ([]uint32, error) {
	var errs error
	for _, d := range f {
		additionalGids, err := d.AdditionalGIDs()
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		return additionalGids, nil
	}
	return nil, nil
}
