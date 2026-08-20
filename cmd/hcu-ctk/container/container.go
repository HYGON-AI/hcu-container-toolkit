/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package container

import (
	"github.com/urfave/cli/v2"
	"hcu-container-toolkit/internal/logger"
	"hcu-container-toolkit/internal/query"
)

type command struct {
	logger logger.Interface
}

type config struct {
	hcus      string
	runtime   string
	namespace string
}

func NewCommand(logger logger.Interface) *cli.Command {
	c := command{
		logger: logger,
	}
	return c.build()
}

func (m command) build() *cli.Command {
	cfg := config{}
	container := cli.Command{
		Name:  "container",
		Usage: "List which containers are using the HCUs",
		Action: func(context *cli.Context) error {
			return m.run(context, &cfg)
		},
	}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:        "hcus",
			Aliases:     []string{"d"},
			Value:       "all",
			Usage:       "Specify the target HCU",
			Destination: &cfg.hcus,
		},
		&cli.StringFlag{
			Name:        "runtime",
			Aliases:     []string{"r"},
			Usage:       "Specify the target runtime engine; one of [docker, containerd, podman]",
			Value:       "docker",
			Destination: &cfg.runtime,
		},
		&cli.StringFlag{
			Name:        "namespace",
			Aliases:     []string{"n"},
			Usage:       "Specify the containerd namespace",
			Value:       "default",
			Destination: &cfg.namespace,
		},
	}
	container.Flags = flags
	return &container
}

func (m command) run(c *cli.Context, cfg *config) error {
	err := query.ShowStatus(cfg.hcus, cfg.runtime, cfg.namespace)
	if err != nil {
		return err
	}
	return nil
}
