/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/pelletier/go-toml"
)

// Toml is a type for the TOML representation of a config.
type Toml struct {
	tree *toml.Tree
	// valuesSet allows us to trac explicitly set values so that we can
	// properly uncomment defaults.
	valuesSet map[string]bool
}

func TreeFromMap(m map[string]any) (*Toml, error) {
	tree, err := toml.TreeFromMap(m)
	if err != nil {
		return nil, err
	}
	return fromTree(tree), nil
}

func fromTree(t *toml.Tree) *Toml {
	return &Toml{tree: t, valuesSet: make(map[string]bool)}
}

type options struct {
	configFile string
	required   bool
}

// Option is a functional option for loading TOML config files.
type Option func(*options)

// WithConfigFile sets the config file option.
func WithConfigFile(configFile string) Option {
	return func(o *options) {
		o.configFile = configFile
	}
}

// WithRequired sets the required option.
// If this is set to true, a failure to open the specified file is treated as an error
func WithRequired(required bool) Option {
	return func(o *options) {
		o.required = required
	}
}

// New creates a new toml tree based on the provided options
func New(opts ...Option) (*Toml, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	return o.loadConfigToml()
}

func (o options) loadConfigToml() (*Toml, error) {
	filename := o.configFile
	if filename == "" {
		return defaultToml()
	}

	_, err := os.Stat(filename)
	if os.IsNotExist(err) && o.required {
		return nil, os.ErrNotExist
	}

	tomlFile, err := os.Open(filename)
	if os.IsNotExist(err) {
		return defaultToml()
	} else if err != nil {
		return nil, fmt.Errorf("failed to load specified config file: %w", err)
	}
	defer tomlFile.Close()

	t, err := loadConfigTomlFrom(tomlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load specified config file: %w", err)
	}

	for _, key := range t.tree.Keys() {
		t.markKeysSet(key, "")
	}
	return t, nil
}

func (t *Toml) markKeysSet(key string, prefix string) {
	fullKey := key
	if prefix != "" {
		fullKey = prefix + "." + key
	}

	t.valuesSet[fullKey] = true

	subTree := t.tree.Get(fullKey)
	if nextTree, ok := subTree.(*toml.Tree); ok {
		for _, nextKey := range nextTree.Keys() {
			t.markKeysSet(nextKey, fullKey)
		}
	}
}

func defaultToml() (*Toml, error) {
	cfg, err := GetDefault()
	if err != nil {
		return nil, err
	}
	contents, err := toml.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	return loadConfigTomlFrom(bytes.NewReader(contents))
}

func loadConfigTomlFrom(reader io.Reader) (*Toml, error) {
	tree, err := toml.LoadReader(reader)
	if err != nil {
		return nil, err
	}

	t := &Toml{
		tree:      tree,
		valuesSet: make(map[string]bool),
	}

	return t, nil
}

// Config returns the typed config associated with the toml tree.
func (t *Toml) Config() (*Config, error) {
	cfg, err := GetDefault()
	if err != nil {
		return nil, err
	}
	if t == nil {
		return cfg, nil
	}
	if err := t.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}
	return cfg, nil
}

// Unmarshal wraps the toml.Tree Unmarshal function.
func (t *Toml) Unmarshal(v interface{}) error {
	return t.tree.Unmarshal(v)
}

// Save saves the config to the specified Writer.
func (t *Toml) Save(w io.Writer) (int64, error) {
	contents, err := t.contents()
	if err != nil {
		return 0, err
	}

	n, err := w.Write(contents)
	return int64(n), err
}

// contents returns the config TOML as a byte slice.
// Any required formatting is applied.
func (t Toml) contents() ([]byte, error) {
	commented := t.commentDefaults()

	buffer := bytes.NewBuffer(nil)

	enc := toml.NewEncoder(buffer).Indentation("")
	if err := enc.Encode(commented.tree); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}
	return t.format(buffer.Bytes())
}

// format fixes the comments for the config to ensure that they start in column
// 1 and are not followed by a space.
func (t Toml) format(contents []byte) ([]byte, error) {
	r := regexp.MustCompile(`(\n*)\s*?#\s*(\S.*)`)
	replaced := r.ReplaceAll(contents, []byte("$1#$2"))

	return replaced, nil
}

// Delete deletes the specified key from the TOML config.
func (t *Toml) Delete(key string) error {
	delete(t.valuesSet, key)
	return t.tree.Delete(key)
}

// Get returns the value for the specified key.
func (t *Toml) Get(key string) interface{} {
	return t.tree.Get(key)
}

// Set sets the specified key to the specified value in the TOML config.
func (t *Toml) Set(key string, value interface{}) {
	t.tree.Set(key, value)
	t.valuesSet[key] = true
}

// commentDefaults applies the required comments for default values to the Toml.
func (t *Toml) commentDefaults() *Toml {
	asToml := t.tree
	commentedDefaults := map[string]interface{}{
		"swarm-resource": "DOCKER_RESOURCE_HCU",
		"accept-hcu-visible-devices-envvar-when-unprivileged": true,
		"accept-hcu-visible-devices-as-volume-mounts":         false,
		"hcu-container-runtime.debug":                         "/var/log/hcu-container-runtime.log",
	}
	for k, v := range commentedDefaults {
		// If a value has been explicitly set, we don't check whether it should
		// be commented.
		if t.valuesSet[k] {
			continue
		}
		set := asToml.Get(k)
		if !shouldComment(k, v, set) {
			continue
		}
		asToml.SetWithComment(k, "", true, v)
	}
	t.tree = asToml
	return t
}

func shouldComment(key string, defaultValue interface{}, setTo interface{}) bool {
	if key == "hcu-container-runtime.debug" && setTo == "/dev/null" {
		return true
	}
	if setTo == nil || defaultValue == setTo || setTo == "" {
		return true
	}
	return false
}
