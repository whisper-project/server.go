/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package saywhat

import (
	"github.com/gin-gonic/gin"

	"clickonetwo.io/whisper/internal/middleware"
)

func AddRoutes(r *gin.RouterGroup) {
	r.GET("/settings/:profileId", settingsNotImplemented)
	r.POST("/settings/:profileId", settingsNotImplemented)
	r.PUT("/settings/:profileId", settingsNotImplemented)
}

func settingsNotImplemented(c *gin.Context) {
	profileId := c.Param("profileId")
	method := c.Request.Method
	middleware.CtxLogS(c).Errorw("Not Implemented: %s of settings %s", method, profileId)
	c.JSON(500, gin.H{"status": "error", "error": "not implemented: server is under construction"})
}
