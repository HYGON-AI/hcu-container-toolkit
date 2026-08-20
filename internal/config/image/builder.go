/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package image

import (
	"fmt"
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
)

type builder struct {
	env    map[string]string
	mounts []specs.Mount
}

// New creates a new DTK image from the input options.
func New(opt ...Option) (DTK, error) {
	b := &builder{}
	for _, o := range opt {
		if err := o(b); err != nil {
			return DTK{}, err
		}
	}
	if b.env == nil {
		b.env = make(map[string]string)
	}

	return b.build()
}

// build creates a DTK image from the builder.
func (b builder) build() (DTK, error) {
	c := DTK{
		env:    b.env,
		mounts: b.mounts,
	}
	return c, nil
}

// Option is a functional option for creating a DTK image.
type Option func(*builder) error

// WithEnv sets the environment variables to use when creating the DTK image.
// Note that this also overwrites the values set with WithEnvMap.
func WithEnv(envs []string) Option {
	return func(b *builder) error {
		envmap := make(map[string]string)
		for _, env := range envs {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid environment variables: %v", env)
			}
			envmap[parts[0]] = parts[1]
		}
		return WithEnvMap(envmap)(b)
	}
}

// WithEnvMap sets the environment variable map to use when creating the DTK image.
// Note that this also overwrites the values set with WithEnv.
func WithEnvMap(env map[string]string) Option {
	return func(b *builder) error {
		b.env = env
		return nil
	}
}

// WithMounts sets the mounts associated with the DTK image.
func WithMounts(mounts []specs.Mount) Option {
	return func(b *builder) error {
		b.mounts = mounts
		return nil
	}
}
