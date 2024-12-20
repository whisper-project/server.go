/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package console

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/whisper-project/server.go/internal/middleware"
	"github.com/whisper-project/server.go/internal/storage"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type WhisperConversationMap string

func (p WhisperConversationMap) StoragePrefix() string {
	return "whisper-conversations"
}

func (p WhisperConversationMap) StorageId() string {
	return string(p)
}

func GetProfileWhisperConversations(c *gin.Context) {
	profileId := c.Param("profileId")
	if profileId == "" {
		middleware.CtxLog(c).Error("missing profileId in whisper conversations request")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "missing profileId"})
		return
	}
	key := WhisperConversationMap(profileId)
	cMap, err := storage.MapGetAll(c.Request.Context(), key)
	if err != nil {
		middleware.CtxLog(c).Error("storage error on whisper conversations retrieval",
			zap.String("profileId", profileId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	middleware.CtxLog(c).Debug("retrieved whisper conversations",
		zap.String("profileId", profileId), zap.Any("conversations", cMap))
	c.JSON(http.StatusOK, cMap)
}

func PostProfileWhisperConversation(c *gin.Context) {
	profileId := c.Param("profileId")
	if profileId == "" {
		middleware.CtxLog(c).Error("missing profileId in whisper conversations request")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "missing profileId"})
		return
	}
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
	key := WhisperConversationMap(profileId)
	cMap, err := storage.MapGetAll(c.Request.Context(), key)
	if err != nil {
		middleware.CtxLog(c).Error("storage error on whisper conversations retrieval",
			zap.String("profileId", profileId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	if _, ok := cMap[name]; ok {
		middleware.CtxLog(c).Error("whisper conversation already exists",
			zap.String("profileId", profileId), zap.String("name", name))
		c.JSON(http.StatusConflict, gin.H{"status": "error", "error": "whisper conversation already exists"})
	}
	conversationId := uuid.NewString()
	if err := storage.MapSet(c.Request.Context(), key, name, conversationId); err != nil {
		middleware.CtxLog(c).Error("storage error on whisper conversation creation",
			zap.String("profileId", profileId), zap.String("name", name), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, conversationId)
}
