/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package console

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProfileWhisperConversationMap(c *gin.Context) {
	mock := map[string]string{"mock": "mock-conversation-id"}
	c.JSON(http.StatusOK, mock)
}

func PostProfileWhisperConversationMap(c *gin.Context) {
	c.JSON(http.StatusCreated, "mock-conversation-id")
}
