/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package console

import (
	"github.com/gin-gonic/gin"
	"github.com/whisper-project/server.golang/common/profile"
)

func AddRoutes(r *gin.RouterGroup) {
	r.POST("/preferences", PostPrefsHandler)
	r.POST("/request-email", PostRequestEmailHandler)
	r.GET("/profiles/:profileId/whisper-conversations", profile.GetProfileWhisperConversations)
	r.POST("/profiles/:profileId/whisper-conversations", profile.PostProfileWhisperConversation)
	r.GET("/profiles/:profileId/whisper-conversations/:name", profile.GetProfileWhisperConversationId)
	r.DELETE("/profiles/:profileId/whisper-conversations/:name", profile.DeleteProfileWhisperConversation)
}
