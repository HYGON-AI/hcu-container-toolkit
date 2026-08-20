/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

import "sync"

type cache struct {
	d Discover

	sync.Mutex
	devices      []Device
	hooks        []Hook
	mounts       []Mount
	additionGids []uint32
}

var _ Discover = (*cache)(nil)

// WithCache decorates the specified disoverer with a cache.
func WithCache(d Discover) Discover {
	if d == nil {
		return None{}
	}
	return &cache{d: d}
}

func (c *cache) Devices() ([]Device, error) {
	c.Lock()
	defer c.Unlock()

	if c.devices == nil {
		devices, err := c.d.Devices()
		if err != nil {
			return nil, err
		}
		c.devices = devices
	}
	return c.devices, nil
}

func (c *cache) Hooks() ([]Hook, error) {
	c.Lock()
	defer c.Unlock()

	if c.hooks == nil {
		hooks, err := c.d.Hooks()
		if err != nil {
			return nil, err
		}
		c.hooks = hooks
	}
	return c.hooks, nil
}

func (c *cache) Mounts() ([]Mount, error) {
	c.Lock()
	defer c.Unlock()

	if c.mounts == nil {
		mounts, err := c.d.Mounts()
		if err != nil {
			return nil, err
		}
		c.mounts = mounts
	}
	return c.mounts, nil
}

func (c *cache) AdditionalGIDs() ([]uint32, error) {
	c.Lock()
	defer c.Unlock()

	if c.additionGids == nil {
		additionGids, err := c.d.AdditionalGIDs()
		if err != nil {
			return nil, err
		}
		c.additionGids = additionGids
	}
	return c.additionGids, nil
}
