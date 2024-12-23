/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package profile

import (
	"fmt"
	"net/http"

	"github.com/whisper-project/server.golang/common/middleware"

	"github.com/whisper-project/server.golang/common/auth"

	"github.com/whisper-project/server.golang/common/storage"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

func AuthenticateRequest(c *gin.Context, profileId string) *Profile {
	if profileId == "" {
		middleware.CtxLog(c).Info("missing profileId in request")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "missing profileId"})
		return nil
	}
	p := &Profile{Id: profileId}
	if err := storage.LoadFields(c.Request.Context(), p); err != nil {
		middleware.CtxLog(c).Error("Load Fields failure", zap.String("profileId", profileId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil
	}
	if auth.AuthenticateRequest(c, profileId, p.Secret) {
		return p
	}
	return nil
}

func NewProfile(c *gin.Context, hashedEmail string) (*Profile, error) {
	ctx := c.Request.Context()
	p := &Profile{
		Id:        uuid.NewString(),
		EmailHash: hashedEmail,
		Secret:    uuid.NewString(),
	}
	if err := storage.SaveFields(ctx, p); err != nil {
		middleware.CtxLog(c).Error("Save Fields failure",
			zap.String("profileId", p.Id), zap.Error(err))
		return nil, err
	}
	if err := storage.MapSet(ctx, EmailProfileMap, hashedEmail, p.Id); err != nil {
		middleware.CtxLog(c).Error("Map set failure",
			zap.String("email", hashedEmail), zap.String("profileId", p.Id), zap.Error(err))
		return nil, err
	}
	if _, err := AddWhisperConversation(c, p.Id, "Conversation 1"); err != nil {
		return nil, err
	}
	return p, nil
}

func GetProfileWhisperConversations(c *gin.Context) {
	profileId := c.Param("profileId")
	if AuthenticateRequest(c, profileId) == nil {
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
	middleware.CtxLog(c).Info("retrieved whisper conversations",
		zap.String("profileId", profileId), zap.String("clientId", c.GetHeader("X-Client-Id")),
		zap.Int("count", len(cMap)))
	middleware.CtxLog(c).Debug("retrieved whisper conversations",
		zap.String("profileId", profileId), zap.String("clientId", c.GetHeader("X-Client-Id")),
		zap.Any("conversations", cMap))
	c.JSON(http.StatusOK, cMap)
}

func GetProfileWhisperConversationId(c *gin.Context) {
	profileId := c.Param("profileId")
	name := c.Param("name")
	if name == "" {
		middleware.CtxLog(c).Info("empty whisper conversation name", zap.String("profileId", profileId))
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "empty whisper conversation name"})
		return
	}
	if AuthenticateRequest(c, profileId) == nil {
		return
	}
	conversationId, err := WhisperConversation(c, profileId, c.Param("name"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	if conversationId == "" {
		middleware.CtxLog(c).Info("whisper conversation not found",
			zap.String("profileId", profileId), zap.String("name", c.Param("name")))
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": "whisper conversation not found"})
	}
	middleware.CtxLog(c).Info("retrieved whisper conversation",
		zap.String("profileId", profileId), zap.String("clientId", c.GetHeader("X-Client-Id")),
		zap.String("name", c.Param("name")), zap.String("conversationId", conversationId))
	c.JSON(http.StatusOK, conversationId)
}

func PostProfileWhisperConversation(c *gin.Context) {
	profileId := c.Param("profileId")
	if AuthenticateRequest(c, profileId) == nil {
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
	cMap, err := WhisperConversations(c, profileId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	if _, ok := cMap[name]; ok {
		middleware.CtxLog(c).Info("whisper conversation already exists",
			zap.String("profileId", profileId), zap.String("name", name))
		c.JSON(http.StatusConflict, gin.H{"status": "error", "error": "whisper conversation already exists"})
		return
	}
	conversationId, err := AddWhisperConversation(c, profileId, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	middleware.CtxLog(c).Info("created whisper conversation",
		zap.String("profileId", profileId), zap.String("clientId", c.GetHeader("X-Client-Id")),
		zap.String("name", c.Param("name")), zap.String("conversationId", conversationId))
	c.JSON(http.StatusCreated, conversationId)
}

func DeleteProfileWhisperConversation(c *gin.Context) {
	profileId := c.Param("profileId")
	if AuthenticateRequest(c, profileId) == nil {
		return
	}
	name := c.Param("name")
	if name == "" {
		middleware.CtxLog(c).Error("empty whisper conversation name")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "empty whisper conversation name"})
		return
	}
	conversationId, err := WhisperConversation(c, profileId, name)
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
	if err := DeleteWhisperConversation(c, profileId, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func WhisperConversation(c *gin.Context, profileId string, name string) (string, error) {
	if name == "" {
		return "", nil
	}
	key := WhisperConversationMap(profileId)
	conversationId, err := storage.MapGet(c.Request.Context(), key, name)
	if err != nil {
		middleware.CtxLog(c).Error("storage error on whisper conversation retrieval",
			zap.String("profileId", profileId), zap.String("name", c.Param("name")), zap.Error(err))
		return "", err
	}
	return conversationId, nil
}

func WhisperConversations(c *gin.Context, profileId string) (map[string]string, error) {
	key := WhisperConversationMap(profileId)
	cMap, err := storage.MapGetAll(c.Request.Context(), key)
	if err != nil {
		middleware.CtxLog(c).Error("storage error on whisper conversations retrieval",
			zap.String("profileId", profileId), zap.Error(err))
		return nil, err
	}
	return cMap, nil
}

func AddWhisperConversation(c *gin.Context, profileId string, name string) (string, error) {
	key := WhisperConversationMap(profileId)
	conversationId := uuid.NewString()
	if err := storage.MapSet(c.Request.Context(), key, name, conversationId); err != nil {
		middleware.CtxLog(c).Error("storage error on whisper conversation creation",
			zap.String("profileId", profileId), zap.String("name", name), zap.Error(err))
		return "", err
	}
	return conversationId, nil
}

func DeleteWhisperConversation(c *gin.Context, profileId string, name string) error {
	key := WhisperConversationMap(profileId)
	if err := storage.MapRemove(c.Request.Context(), key, name); err != nil {
		middleware.CtxLog(c).Error("storage error on whisper conversation creation",
			zap.String("profileId", profileId), zap.String("name", name), zap.Error(err))
		return err
	}
	return nil
}

func EmailProfile(c *gin.Context, hashedEmail string) (string, error) {
	profileId, err := storage.MapGet(c.Request.Context(), EmailProfileMap, hashedEmail)
	if err != nil {
		middleware.CtxLog(c).Error("EmailProfileMap failure", zap.String("email", hashedEmail), zap.Error(err))
		return "", err
	}
	return profileId, nil
}

func SetEmailProfile(c *gin.Context, hashedEmail, profileId string) error {
	if err := storage.MapSet(c.Request.Context(), EmailProfileMap, hashedEmail, profileId); err != nil {
		middleware.CtxLog(c).Error("EmailProfileMap set failure",
			zap.String("email", hashedEmail), zap.String("profileId", profileId), zap.Error(err))
		return err
	}
	return nil
}

func ClientProfile(c *gin.Context, clientId string) (string, error) {
	profileId, err := storage.MapGet(c.Request.Context(), ClientProfileMap, clientId)
	if err != nil {
		middleware.CtxLog(c).Error("ClientProfileMap failure", zap.String("clientId", clientId), zap.Error(err))
		return "", err
	}
	return profileId, nil
}

func SetClientProfile(c *gin.Context, clientId, profileId string) error {
	if err := storage.MapSet(c.Request.Context(), ClientProfileMap, clientId, profileId); err != nil {
		middleware.CtxLog(c).Error("ClientProfileMap set failure",
			zap.String("clientId", clientId), zap.String("profileId", profileId), zap.Error(err))
		return err
	}
	return nil
}
