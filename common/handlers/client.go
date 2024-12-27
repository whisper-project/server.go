/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/whisper-project/server.golang/common/platform"
	"github.com/whisper-project/server.golang/common/storage"
)

func ObserveClientLaunch(c *gin.Context, clientId, profileId string) {
	l := storage.NewLaunchData(clientId, profileId)
	_ = platform.SaveFields(c.Request.Context(), l)
}

func ObserveClientShutdown(c *gin.Context, clientId string) {
	l := &storage.LaunchData{ClientId: clientId}
	if err := platform.LoadFields(c.Request.Context(), l); err != nil {
		return
	}
	l.End = time.Now()
	_ = platform.SaveFields(c.Request.Context(), l)
}

func ObserveClientIdle(c *gin.Context, clientId string) {
	l := &storage.LaunchData{ClientId: clientId}
	if err := platform.LoadFields(c.Request.Context(), l); err != nil {
		return
	}
	l.Idle = time.Now()
	_ = platform.SaveFields(c.Request.Context(), l)
}
