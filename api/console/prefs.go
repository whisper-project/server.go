/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package console

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/whisper-project/server.golang/common/middleware"

	"github.com/whisper-project/server.golang/common/profile"

	"github.com/whisper-project/server.golang/common/storage"

	"gopkg.in/gomail.v2"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func PostPrefsHandler(c *gin.Context) {
	clientId := c.GetHeader("X-Client-Id")
	if clientId == "" {
		middleware.CtxLog(c).Info("Missing X-Client-Id header")
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	var emailHash string
	if err := c.ShouldBindJSON(&emailHash); err != nil {
		middleware.CtxLog(c).Info("Invalid body", zap.Error(err))
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	// Look for a profile that matches the email
	profileId, err := profile.EmailProfile(c, emailHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if profileId != "" {
		// Found an existing profile with that email
		if c.GetHeader("Authorization") == "" {
			// User needs to provide password to authorize against this profile, so challenge with it
			middleware.CtxLog(c).Info("Profile exists, need authorization", zap.String("profileId", profileId))
			c.JSON(http.StatusUnauthorized, profileId)
			return
		}
		// Check the user's authorization
		p := profile.AuthenticateRequest(c, profileId)
		if p == nil {
			return
		}
		// they are authorized, so remember them against this client
		if err = profile.SetClientProfile(c, clientId, p.Id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		middleware.CtxLog(c).Info("Authenticated profile",
			zap.String("email", p.EmailHash), zap.String("profileId", p.Id), zap.String("clientId", clientId))
		c.Status(http.StatusNoContent)
		return
	}
	// This is a new email, generate a profile for it, and record it against email and client
	p, err := profile.NewProfile(c, emailHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err = profile.SetEmailProfile(c, p.EmailHash, p.Id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err = profile.SetClientProfile(c, clientId, p.Id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.CtxLog(c).Info("Created new profile",
		zap.String("email", p.EmailHash), zap.String("profileId", p.Id), zap.String("clientId", clientId))
	response := map[string]string{"id": p.Id, "secret": p.Secret}
	c.JSON(http.StatusCreated, response)
}

func PostRequestEmailHandler(c *gin.Context) {
	var email string
	err := c.Bind(&email)
	if err != nil || email == "" {
		middleware.CtxLog(c).Error("Invalid request for email", zap.String("email", email), zap.Error(err))
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	hash := makeSha1(email)

	// Look for a profile that matches the email
	ctx := c.Request.Context()
	profileId, err := storage.MapGet(ctx, profile.EmailProfileMap, hash)
	if err != nil {
		middleware.CtxLog(c).Error("Map failure", zap.String("email", hash), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// if we don't find one, let the client know
	if profileId == "" {
		middleware.CtxLog(c).Error("No profile found for email", zap.String("email", hash))
		c.JSON(http.StatusNotFound, gin.H{"error": "No profile found for email"})
		return
	}
	// otherwise, load the profile, and send email with password
	p := &profile.Profile{Id: profileId}
	if err := storage.LoadFields(ctx, p); err != nil {
		middleware.CtxLog(c).Error("Load Fields failure", zap.String("profileId", profileId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	middleware.CtxLog(c).Info("Sending email", zap.String("profileId", p.Id), zap.String("password", p.Secret))
	if err := sendMail(email, p.Secret); err != nil {
		middleware.CtxLog(c).Error("Send email failure", zap.String("profileId", p.Id), zap.Error(err))
	}
	c.Status(http.StatusNoContent)
}

// from https://stackoverflow.com/a/10701951/558006
func makeSha1(s string) string {
	hashFn := sha1.New()
	hashFn.Write([]byte(s))
	return base64.URLEncoding.EncodeToString(hashFn.Sum(nil))
}

func sendMail(to, pw string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "no-reply@whisper-project.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Your Whisper profile information")
	m.SetBody("text/html", "As requested, your Whisper profile password is: <pre>"+pw+"</pre>")

	account := os.Getenv("SMTP_ACCOUNT")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port, err := strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 16)
	if err != nil || account == "" || password == "" || host == "" {
		return fmt.Errorf("missing SMTP environment variables")
	}
	d := gomail.NewDialer(host, int(port), account, password)

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
