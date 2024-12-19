/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package console

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/whisper-project/server.go/internal/middleware"
	"github.com/whisper-project/server.go/internal/storage"

	client "github.com/whisper-project/client.go/api"
)

var (
	EmailProfileMap  = StoredMap("email_profile_map")
	ClientProfileMap = StoredMap("client_profile_map")
)

type Profile struct {
	Id       string `redis:"id"`
	Email    string `redis:"email"`
	Password string `redis:"password"`
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

func NewProfile(ctx context.Context, email string) (*Profile, error) {
	p := &Profile{
		Id:       uuid.NewString(),
		Email:    email,
		Password: uuid.NewString(),
	}
	if err := storage.SaveFields(ctx, p); err != nil {
		return nil, err
	}
	if err := storage.MapSet(ctx, EmailProfileMap, email, p.Id); err != nil {
		return nil, err
	}
	return p, nil
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

type StoredMap string

func (s StoredMap) StoragePrefix() string {
	return "map:"
}

func (s StoredMap) StorageId() string {
	return string(s)
}

func PostPrefsHandler(c *gin.Context) {
	var req client.Prefs
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Look for a profile that matches the email
	ctx := c.Request.Context()
	profileId, err := storage.MapGet(ctx, EmailProfileMap, req.ProfileEmail)
	if err != nil {
		middleware.CtxLog(c).Error("Map failure", zap.String("email", req.ProfileEmail), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Found an existing profile with that email, so check password
	if profileId != "" {
		// got a profile, check that the password is right
		p := &Profile{Id: profileId}
		if err = storage.LoadFields(ctx, p); err != nil {
			middleware.CtxLog(c).Error("Load Fields failure", zap.String("profileId", p.Id), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if p.Password == req.ProfileSecret {
			middleware.CtxLog(c).Info("Found matching profile", zap.String("email", p.Email), zap.String("profileId", p.Id))
			if err = storage.MapSet(ctx, ClientProfileMap, req.ClientId, p.Id); err != nil {
				middleware.CtxLog(c).Error("Map set failure", zap.String("clientId", req.ClientId), zap.String("profileId", p.Id), zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, req)
			return
		}
		middleware.CtxLog(c).Info("Profile password mismatch", zap.String("email", p.Email), zap.String("profileId", p.Id))
		c.JSON(409, gin.H{"error": "Invalid password"})
		return
	}

	// Generate a new profile and return it
	p, err := NewProfile(ctx, req.ProfileEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.CtxLog(c).Info("Created new profile", zap.String("email", p.Email), zap.String("profileId", p.Id))
	if err = storage.MapSet(ctx, ClientProfileMap, req.ClientId, p.Id); err != nil {
		middleware.CtxLog(c).Error("Map set failure", zap.String("clientId", req.ClientId), zap.String("profileId", p.Id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req.ProfileId = p.Id
	req.ProfileSecret = p.Password
	c.JSON(201, req)
}
