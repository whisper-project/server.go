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

	"gopkg.in/gomail.v2"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/whisper-project/server.golang/internal/middleware"
	"github.com/whisper-project/server.golang/internal/storage"

	client "github.com/whisper-project/client.golang/api"
)

func PostPrefsHandler(c *gin.Context) {
	var req client.Prefs
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	// Look for a profile that matches the email
	profileId, err := EmailProfile(c, req.ProfileEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if profileId != "" {
		// Found an existing profile with that email
		if c.GetHeader("Authorization") == "" {
			// User needs to provide password to authorize against this profile, so challenge with it
			middleware.CtxLog(c).Info("Profile exists, need authorization", zap.String("profileId", profileId))
			req.ProfileId = profileId
			c.JSON(http.StatusUnauthorized, req)
			return
		}
		// Check the user's authorization
		p := AuthenticateRequest(c, profileId)
		if p == nil {
			return
		}
		// they are authorized, so remember them against this client
		if err = SetClientProfile(c, req.ClientId, p.Id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		middleware.CtxLog(c).Info("Authenticated profile",
			zap.String("email", p.EmailHash), zap.String("profileId", p.Id), zap.String("clientId", req.ClientId))
		c.Status(http.StatusNoContent)
		return
	}
	// This is a new email, generate a profile for it, and record it against email and client
	p, err := NewProfile(c, req.ProfileEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err = SetEmailProfile(c, p.EmailHash, p.Id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err = SetClientProfile(c, req.ClientId, p.Id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.CtxLog(c).Info("Created new profile",
		zap.String("email", p.EmailHash), zap.String("profileId", p.Id), zap.String("clientId", req.ClientId))
	req.ProfileId = p.Id
	req.ProfileSecret = p.Password
	c.JSON(http.StatusCreated, req)
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
	profileId, err := storage.MapGet(ctx, EmailProfileMap, hash)
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
	p := &Profile{Id: profileId}
	if err := storage.LoadFields(ctx, p); err != nil {
		middleware.CtxLog(c).Error("Load Fields failure", zap.String("profileId", profileId), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	middleware.CtxLog(c).Info("Sending email", zap.String("profileId", p.Id), zap.String("password", p.Password))
	if err := sendMail(email, p.Password); err != nil {
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
