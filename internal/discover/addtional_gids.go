/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright (c) 2026 Hygon Information Technology Co., Ltd.
 */

package discover

import (
	"hcu-container-toolkit/internal/info"
	"hcu-container-toolkit/internal/logger"
	"os/user"
	"strconv"
)

type userGroup struct {
	None
	logger logger.Interface
	groups []string
}

var _ Discover = (*userGroup)(nil)

// NewUserGroupDiscover
func NewUserGroupDiscover(logger logger.Interface, groups ...string) Discover {
	if len(groups) == 0 {
		grps, err := info.GetAdditionalGroups()
		if err != nil {
			logger.Warningf("failed to get groups: %v", err)
		}
		groups = append(groups, grps...)
	}
	return &userGroup{
		logger: logger,
		groups: groups,
	}
}

func (g *userGroup) AdditionalGIDs() ([]uint32, error) {
	var gids []uint32

	for _, group := range g.groups {
		gid, err := GetGid(group)
		if err != nil {
			g.logger.Warningf("Failed to get group id: %s, %v", gid, err)
			continue
		}
		gids = append(gids, uint32(gid))
	}

	return gids, nil
}

func GetGid(group string) (int, error) {
	g, err := user.LookupGroup(group)
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(g.Gid)
}
