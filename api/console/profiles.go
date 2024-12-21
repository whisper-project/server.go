/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package console

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/whisper-project/server.golang/internal/middleware"
	"github.com/whisper-project/server.golang/internal/storage"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type StoredMap string

func (s StoredMap) StoragePrefix() string {
	return "map:"
}

func (s StoredMap) StorageId() string {
	return string(s)
}

var (
	EmailProfileMap  = StoredMap("email_profile_map")
	ClientProfileMap = StoredMap("client_profile_map")
)

type Profile struct {
	Id        string `redis:"id"`
	EmailHash string `redis:"emailHash"`
	Password  string `redis:"password"`
}

func (p *Profile) StoragePrefix() string {
	return "profile:"
}

func (p *Profile) StorageId() string {
	if p == nil {
		return ""
	}
	return p.Id
}

func (p *Profile) SetStorageId(id string) error {
	if p == nil {
		return fmt.Errorf("can't set storage id of nil struct")
	}
	p.Id = id
	return nil
}

func (p *Profile) Copy() storage.StructPointer {
	if p == nil {
		return nil
	}
	n := new(Profile)
	*n = *p
	return n
}

func (p *Profile) Downgrade(in any) (storage.StructPointer, error) {
	if o, ok := in.(Profile); ok {
		return &o, nil
	}
	if o, ok := in.(*Profile); ok {
		return o, nil
	}
	return nil, fmt.Errorf("not a %T: %#v", p, in)
}

func NewProfile(c *gin.Context, hashedEmail string) (*Profile, error) {
	ctx := c.Request.Context()
	p := &Profile{
		Id:        uuid.NewString(),
		EmailHash: hashedEmail,
		Password:  uuid.NewString(),
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
	if _, err := AddWhisperConversation(c, p.Id, "Conversation-1"); err != nil {
		return nil, err
	}
	return p, nil
}

type WhisperConversationMap string

func (p WhisperConversationMap) StoragePrefix() string {
	return "whisper-conversations"
}

func (p WhisperConversationMap) StorageId() string {
	return string(p)
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

func GetProfileWhisperConversation(c *gin.Context) {
	profileId := c.Param("profileId")
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
