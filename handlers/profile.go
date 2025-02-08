/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package handlers

import (
	"fmt"
	"net/http"

	"github.com/whisper-project/server.golang/middleware"
	"github.com/whisper-project/server.golang/platform"
	"github.com/whisper-project/server.golang/storage"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

func PatchProfileHandler(c *gin.Context) {
	var updates map[string]string
	if err := c.ShouldBindJSON(&updates); err != nil {
		middleware.CtxLog(c).Info("Can't bind patch map", zap.Error(err))
		c.JSON(400, gin.H{"error": "Invalid request format"})
		return
	}
	p := AuthenticateRequest(c)
	if p == nil {
		return
	}
	updated := false
	if n, ok := updates["name"]; ok && n != "" && n != p.Name {
		p.Name = n
		updated = true
	}
	if s, ok := updates["secret"]; ok && s != "" && s != p.Secret {
		p.Secret = s
		updated = true
	}
	if updated {
		if err := platform.SaveFields(c.Request.Context(), p); err != nil {
			middleware.CtxLog(c).Info("Can't save fields on profile patch",
				zap.String("profileId", p.Id), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	middleware.CtxLog(c).Info("Patched profile",
		zap.String("profileId", p.Id), zap.Any("updates", updates))
	c.Status(http.StatusNoContent)
}

func GetProfileWhisperConversationsHandler(c *gin.Context) {
	if AuthenticateRequest(c) == nil {
		return
	}
	profileId := c.GetHeader("X-Profile-Id")
	cMap, err := storage.WhisperConversations(profileId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	middleware.CtxLog(c).Info("retrieved whisper conversations",
		zap.String("profileId", profileId), zap.String("clientId", c.GetHeader("X-Client-Id")),
		zap.Int("count", len(cMap)))
	middleware.CtxLog(c).Debug("retrieved whisper conversations",
		zap.String("profileId", profileId), zap.String("clientId", c.GetHeader("X-Client-Id")),
		zap.Any("conversations", cMap))
	c.JSON(http.StatusOK, cMap)
}

func GetProfileWhisperConversationIdHandler(c *gin.Context) {
	if AuthenticateRequest(c) == nil {
		return
	}
	profileId := c.GetHeader("X-Profile-Id")
	name := c.Param("name")
	if name == "" {
		middleware.CtxLog(c).Info("empty whisper conversation name")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "empty whisper conversation name"})
		return
	}
	conversationId, err := storage.WhisperConversation(profileId, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	if conversationId == "" {
		middleware.CtxLog(c).Info("whisper conversation not found",
			zap.String("profileId", profileId), zap.String("name", name))
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": "whisper conversation not found"})
	}
	middleware.CtxLog(c).Info("retrieved whisper conversation",
		zap.String("profileId", profileId), zap.String("clientId", c.GetHeader("X-Client-Id")),
		zap.String("name", c.Param("name")), zap.String("conversationId", conversationId))
	c.JSON(http.StatusOK, conversationId)
}

func PostProfileWhisperConversationHandler(c *gin.Context) {
	if AuthenticateRequest(c) == nil {
		return
	}
	profileId := c.GetHeader("X-Profile-Id")
	var name string
	err := c.Bind(&name)
	if err != nil {
		middleware.CtxLog(c).Error("error binding whisper conversation name", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
	}
	if name == "" {
		middleware.CtxLog(c).Error("empty whisper conversation name")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "empty whisper conversation name"})
		return
	}
	id, err := storage.WhisperConversation(profileId, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	if id != "" {
		middleware.CtxLog(c).Info("whisper conversation already exists",
			zap.String("profileId", profileId), zap.String("name", name))
		c.JSON(http.StatusConflict, gin.H{"status": "error", "error": "whisper conversation already exists"})
		return
	}
	conversationId, err := storage.AddWhisperConversation(profileId, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	middleware.CtxLog(c).Info("created whisper conversation",
		zap.String("profileId", profileId), zap.String("clientId", c.GetHeader("X-Client-Id")),
		zap.String("name", c.Param("name")), zap.String("conversationId", conversationId))
	c.JSON(http.StatusCreated, conversationId)
}

func DeleteProfileWhisperConversationHandler(c *gin.Context) {
	if AuthenticateRequest(c) == nil {
		return
	}
	profileId := c.GetHeader("X-Profile-Id")
	name := c.Param("name")
	if name == "" {
		middleware.CtxLog(c).Error("empty whisper conversation name")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "empty whisper conversation name"})
		return
	}
	conversationId, err := storage.WhisperConversation(profileId, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	if conversationId == "" {
		middleware.CtxLog(c).Info("whisper conversation not found",
			zap.String("profileId", profileId), zap.String("name", name))
		c.JSON(http.StatusNotFound,
			gin.H{"status": "error", "error": fmt.Sprintf("whisper conversation %q not found", name)})
		return
	}
	if err := storage.DeleteWhisperConversation(profileId, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
