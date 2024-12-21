/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package console

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/whisper-project/server.golang/internal/middleware"
	"github.com/whisper-project/server.golang/internal/storage"
	"go.uber.org/zap"
)

func AuthenticateRequest(c *gin.Context, profileId string) *Profile {
	if profileId == "" {
		middleware.CtxLog(c).Info("missing profileId in request")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "missing profileId"})
		return nil
	}
	clientId := c.GetHeader("X-Client-Id")
	if clientId == "" {
		middleware.CtxLog(c).Info("missing clientId in request")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "missing clientId"})
		return nil
	}
	authToken := c.GetHeader("Authorization")
	if authToken == "" {
		middleware.CtxLog(c).Info("missing Authorization header")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return nil
	} else if len(authToken) > len("Bearer ") {
		authToken = authToken[len("Bearer "):]
	} else {
		middleware.CtxLog(c).Info("invalid Authorization header", zap.String("header", c.GetHeader("Authorization")))
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid authorization header"})
		return nil
	}
	p := &Profile{Id: profileId}
	if err := storage.LoadFields(c.Request.Context(), p); err != nil {
		middleware.CtxLog(c).Error("Load Fields failure", zap.String("profileId", profileId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil
	}
	key, err := uuid.Parse(p.Password)
	if err != nil {
		middleware.CtxLog(c).Error("Password is not a UUID",
			zap.String("profileId", profileId), zap.String("password", p.Password), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil
	}
	keyBytes, err := key.MarshalBinary()
	if err != nil {
		middleware.CtxLog(c).Error("Marshal UUID failure", zap.String("profileId", profileId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return nil
	}
	validator := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			// notest
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return keyBytes, nil
	}
	token, err := jwt.Parse(authToken, validator, jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}))
	if err != nil {
		middleware.CtxLog(c).Info("Invalid bearer token", zap.String("profileId", profileId), zap.Error(err))
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid bearer token"})
		return nil
	}
	if issuer, err := token.Claims.GetIssuer(); err != nil {
		middleware.CtxLog(c).Info("Invalid issuer claim", zap.Error(err))
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid bearer token"})
		return nil
	} else if issuer != clientId {
		middleware.CtxLog(c).Info("Token issuer doesn't match client id",
			zap.String("clientId", clientId), zap.String("issuer", issuer))
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid bearer token"})
		return nil
	}
	if subject, err := token.Claims.GetSubject(); err != nil {
		middleware.CtxLog(c).Info("Invalid subject claim", zap.Error(err))
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid bearer token"})
		return nil
	} else if subject != profileId {
		middleware.CtxLog(c).Info("Token subject doesn't match profile id",
			zap.String("profileId", profileId), zap.String("subject", subject))
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid bearer token"})
		return nil
	}
	if issuedAt, err := token.Claims.GetIssuedAt(); err != nil {
		middleware.CtxLog(c).Info("Invalid issuedAt claim", zap.Error(err))
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid bearer token"})
		return nil
	} else if age := time.Now().Unix() - issuedAt.Unix(); (age < -300) || (age > 300) {
		middleware.CtxLog(c).Info("Token age is too far off",
			zap.String("issuedAt", issuedAt.String()), zap.Int64("age", age))
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid bearer token"})
		return nil
	}
	return p
}
