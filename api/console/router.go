/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package console

import (
	"github.com/gin-gonic/gin"
)

func AddRoutes(r *gin.RouterGroup) {
	r.POST("/preferences", PostPrefsHandler)
	r.GET("/profiles/:profileId/whisper-conversations", GetProfileWhisperConversationMap)
	r.POST("/profiles/:profileId/whisper-conversations", PostProfileWhisperConversationMap)
}
