package auth

import (
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"clickonetwo.io/whisper/server/middleware"
	"clickonetwo.io/whisper/server/storage"
)

func CreateApnsJwt(c *gin.Context) (string, error) {
	config := storage.GetConfig()
	block, _ := pem.Decode([]byte(config.ApnsCredSecret))
	if block == nil || block.Type != "PRIVATE KEY" {
		// notest
		middleware.CtxLogS(c).Errorf("APNS key PEM decode error on: %q", config.ApnsCredSecret)
		return "", fmt.Errorf("APNS key PEM decode error")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		middleware.CtxLogS(c).Errorf("APNS key PKCS8 decode error: %v", err)
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": config.ApnsTeamId,
	})
	token.Header["kid"] = config.ApnsCredId
	signed, err := token.SignedString(key)
	if err != nil {
		middleware.CtxLogS(c).Errorf("APNS key signing error: %v", err)
		return "", err
	}
	return signed, nil
}

func ValidateClientJwt(c *gin.Context, signed, id, secret string) bool {
	key, err := hex.DecodeString(secret)
	if err != nil || len(key) != 32 {
		middleware.CtxLogS(c).Errorf("Server-side secret (%q) is corrupt, can't validate client JWT: %v", err)
		return false
	}
	validator := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			// notest
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	}
	token, err := jwt.Parse(signed, validator, jwt.WithValidMethods([]string{"HS256", "HS384", "HS512"}))
	if err != nil {
		middleware.CtxLogS(c).Errorf("Invalid token: %v", err)
		return false
	}
	if issuer, err := token.Claims.GetIssuer(); err != nil || issuer == "" {
		middleware.CtxLogS(c).Errorf("Token issuer (%q) is invalid: %v,", issuer, err)
		return false
	} else if issuer != id {
		middleware.CtxLogS(c).Errorf("Token issuer (%q) doesn't match client id (%q)", issuer, id)
		return false
	}
	if issuedAt, err := token.Claims.GetIssuedAt(); err != nil || issuedAt == nil {
		middleware.CtxLogS(c).Errorf("Token issued-at (%v) is invalid: %v", issuedAt, err)
		return false
	} else if age := time.Now().Unix() - issuedAt.Unix(); (age < -300) || (age > 300) {
		middleware.CtxLogS(c).Errorf("Client clock is too far off: token age is %d seconds", age)
		return false
	}
	return true
}
