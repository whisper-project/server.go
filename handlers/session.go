/*
 * Copyright 2025 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whisper-project/server.golang/lifecycle"
	"github.com/whisper-project/server.golang/storage"
)

func StartWhisperSessionHandler(c *gin.Context) {
	p := AuthenticateRequest(c)
	if p == nil {
		return
	}
	clientId := c.GetHeader("X-Client-Id")
	conversationId := c.Param("conversationId")
	isOwned, err := storage.IsOwnedConversation(p.Id, conversationId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isOwned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
	}
	s, err := lifecycle.GetSession(conversationId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err = s.AddWhisperer(clientId, p.Id, p.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func StartListenSessionHandler(c *gin.Context) {
	p := AuthenticateRequest(c)
	if p == nil {
		return
	}
	clientId := c.GetHeader("X-Client-Id")
	conversationId := c.Param("conversationId")
	isAllowed, err := storage.IsAllowedListener(p.Id, conversationId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s, err := lifecycle.GetSession(conversationId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	if isAllowed {
		if err = s.AddListener(clientId, p.Id, p.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}
	if err = s.AddListenerRequest(clientId, p.Id, p.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	// TODO: have client wait for what??
}
