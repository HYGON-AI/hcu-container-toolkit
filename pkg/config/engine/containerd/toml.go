/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package containerd

import (
	"fmt"

	"github.com/pelletier/go-toml"
)

// tomlTree is an alias for toml.Tree that allows for extensions.
type tomlTree toml.Tree

func subtreeAtPath(c toml.Tree, path ...string) *tomlTree {
	tree := c.GetPath(path).(*toml.Tree)
	return (*tomlTree)(tree)
}

func (t *tomlTree) insert(other map[string]interface{}) error {

	for key, value := range other {
		if insertsubtree, ok := value.(map[string]interface{}); ok {
			subtree := (*toml.Tree)(t).Get(key).(*toml.Tree)
			return (*tomlTree)(subtree).insert(insertsubtree)
		}
		(*toml.Tree)(t).Set(key, value)
	}
	return nil
}

func (t *tomlTree) applyOverrides(overrides ...map[string]interface{}) error {
	for _, override := range overrides {
		subconfig, err := toml.TreeFromMap(override)
		if err != nil {
			return fmt.Errorf("invalid toml config: %w", err)
		}
		if err := t.insert(subconfig.ToMap()); err != nil {
			return err
		}
	}
	return nil
}
