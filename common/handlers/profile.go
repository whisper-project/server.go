/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package handlers

import (
	"fmt"
	"net/http"

	"github.com/whisper-project/server.golang/common/middleware"
	"github.com/whisper-project/server.golang/common/platform"
	"github.com/whisper-project/server.golang/common/storage"

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
	cMap, err := WhisperConversations(c, profileId)
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
	conversationId, err := WhisperConversation(c, profileId, name)
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
	id, err := WhisperConversation(c, profileId, name)
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

// NewLaunchProfile creates a launch profile for a hashed email from a client and records it in the database.
func NewLaunchProfile(c *gin.Context, hashedEmail, clientId string) (*storage.Profile, error) {
	ctx := c.Request.Context()
	p := storage.NewProfile(hashedEmail)
	if err := platform.SaveFields(ctx, p); err != nil {
		middleware.CtxLog(c).Error("Save Fields failure on new profile creation",
			zap.String("profileId", p.Id), zap.Error(err))
		return nil, err
	}
	cleanup := false
	defer func() {
		if !cleanup {
			return
		}
		_ = platform.MapRemove(ctx, storage.EmailProfileMap, hashedEmail)
		_ = platform.DeleteStorage(ctx, p)
	}()
	if err := SetEmailProfile(c, p.EmailHash, p.Id); err != nil {
		cleanup = true
		return nil, err
	}
	if _, err := AddWhisperConversation(c, p.Id, "Conversation 1"); err != nil {
		cleanup = true
		return nil, err
	}
	ObserveClientLaunch(c, clientId, p.Id)
	return p, nil
}

func WhisperConversation(c *gin.Context, profileId string, name string) (string, error) {
	if name == "" {
		return "", nil
	}
	key := storage.WhisperConversationMap(profileId)
	conversationId, err := platform.MapGet(c.Request.Context(), key, name)
	if err != nil {
		middleware.CtxLog(c).Error("platform error on whisper conversation retrieval",
			zap.String("profileId", profileId), zap.String("name", c.Param("name")), zap.Error(err))
		return "", err
	}
	return conversationId, nil
}

func WhisperConversations(c *gin.Context, profileId string) (map[string]string, error) {
	key := storage.WhisperConversationMap(profileId)
	cMap, err := platform.MapGetAll(c.Request.Context(), key)
	if err != nil {
		middleware.CtxLog(c).Error("platform error on whisper conversations retrieval",
			zap.String("profileId", profileId), zap.Error(err))
		return nil, err
	}
	return cMap, nil
}

func AddWhisperConversation(c *gin.Context, profileId string, name string) (string, error) {
	ctx := c.Request.Context()
	key := storage.WhisperConversationMap(profileId)
	conversation := storage.NewConversation(profileId, name)
	if err := platform.SaveFields(ctx, conversation); err != nil {
		middleware.CtxLog(c).Error("Save Fields failure on whisper conversation creation",
			zap.String("conversationId", conversation.Id), zap.Error(err))
		return "", err
	}
	if err := platform.MapSet(ctx, key, name, conversation.Id); err != nil {
		middleware.CtxLog(c).Error("platform error on whisper conversation creation",
			zap.String("profileId", profileId), zap.String("name", name), zap.Error(err))
		return "", err
	}
	return conversation.Id, nil
}

func DeleteWhisperConversation(c *gin.Context, profileId string, name string) error {
	key := storage.WhisperConversationMap(profileId)
	if err := platform.MapRemove(c.Request.Context(), key, name); err != nil {
		middleware.CtxLog(c).Error("platform error on whisper conversation deletion",
			zap.String("profileId", profileId), zap.String("name", name), zap.Error(err))
		return err
	}
	return nil
}

func EmailProfile(c *gin.Context, hashedEmail string) (string, error) {
	profileId, err := platform.MapGet(c.Request.Context(), storage.EmailProfileMap, hashedEmail)
	if err != nil {
		middleware.CtxLog(c).Error("EmailProfileMap failure", zap.String("email", hashedEmail), zap.Error(err))
		return "", err
	}
	return profileId, nil
}

func SetEmailProfile(c *gin.Context, hashedEmail, profileId string) error {
	if err := platform.MapSet(c.Request.Context(), storage.EmailProfileMap, hashedEmail, profileId); err != nil {
		middleware.CtxLog(c).Error("EmailProfileMap set failure",
			zap.String("email", hashedEmail), zap.String("profileId", profileId), zap.Error(err))
		return err
	}
	return nil
}
