/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package symlinks

import (
	"fmt"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/oci"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

type command struct {
	logger logger.Interface
}

type config struct {
	hostRoot      string
	links         cli.StringSlice
	containerSpec string
}

// NewCommand constructs a hook command with the specified logger
func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

// build
func (m command) build() *cli.Command {
	cfg := config{}

	// Create the '' command
	c := cli.Command{
		Name:  "create-symlinks",
		Usage: "A hook to create symlinks in the container. This can be used to process CSV mount specs",
		Action: func(c *cli.Context) error {
			return m.run(c, &cfg)
		},
	}

	c.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "host-root",
			Usage:       "The root on the host filesystem to use to resolve symlinks",
			Destination: &cfg.hostRoot,
		},
		&cli.StringSliceFlag{
			Name:        "link",
			Usage:       "Specify a specific link to create. The link is specified as target::link",
			Destination: &cfg.links,
		},
		&cli.StringFlag{
			Name:        "container-spec",
			Usage:       "Specify the path to the OCI container spec. If empty or '-' the spec will be read from STDIN",
			Destination: &cfg.containerSpec,
		},
	}

	return &c
}

func (m command) run(c *cli.Context, cfg *config) error {
	s, err := oci.LoadContainerState(cfg.containerSpec)
	if err != nil {
		return fmt.Errorf("failed to load container state: %v", err)
	}

	containerRoot, err := s.GetContainerRoot()
	if err != nil {
		return fmt.Errorf("failed to determined container root: %v", err)
	}

	created := make(map[string]bool)

	links := cfg.links.Value()
	for _, l := range links {
		parts := strings.Split(l, "::")
		if len(parts) != 2 {
			m.logger.Warningf("Invalid link specification %v", l)
			continue
		}

		err := m.createLink(created, cfg.hostRoot, containerRoot, parts[0], parts[1])
		if err != nil {
			m.logger.Warningf("Failed to create link %v: %v", parts, err)
		}
	}

	return nil
}

func DirExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func (m command) createLink(created map[string]bool, hostRoot string, containerRoot string, target string, link string) error {
	linkPath, err := changeRoot(hostRoot, containerRoot, link)
	if err != nil {
		m.logger.Warningf("Failed to resolve path for link %v relative to %v: %v", link, containerRoot, err)
	}
	if created[linkPath] {
		m.logger.Debugf("Link %v already created", linkPath)
		return nil
	}

	targetPath, err := changeRoot(hostRoot, "/", target)
	if err != nil {
		m.logger.Warningf("Failed to resolve path for target %v relative to %v: %v", target, "/", err)
	}

	m.logger.Infof("Symlinking %v to %v", linkPath, targetPath)
	if DirExists(linkPath) {
		os.RemoveAll(linkPath)
		m.logger.Infof("remove dir %v", linkPath)
	}
	err = os.MkdirAll(filepath.Dir(linkPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	err = os.Symlink(target, linkPath)
	if err != nil {
		return fmt.Errorf("failed to create symlink: %v", err)
	}

	return nil
}

func changeRoot(current string, new string, path string) (string, error) {
	if !filepath.IsAbs(path) {
		return path, nil
	}

	relative := path
	if current != "" {
		r, err := filepath.Rel(current, path)
		if err != nil {
			return "", err
		}
		relative = r
	}

	return filepath.Join(new, relative), nil
}
