/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	"github.com/whisper-project/server.go/internal/client"
	"github.com/whisper-project/server.go/internal/middleware"
	"github.com/whisper-project/server.go/internal/storage"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestApnsJwt(t *testing.T) {
	pubKeyPem := "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEmVNsF3M0Y5pZHByaeMh05dypQF6A\nfKMzPCOmyQWT1BpU3SKb1drtpi4SgWTWSFA2qnRypH7pEp/oYHWbKTLWgA==\n-----END PUBLIC KEY-----"
	block, _ := pem.Decode([]byte(pubKeyPem))
	if block == nil || block.Type != "PUBLIC KEY" {
		t.Fatalf("APNS pubkey PEM decode error")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("ParsePKIXPublicKey failed: %v", err)
	}
	pubkey, ok := key.(*ecdsa.PublicKey)
	if !ok {
		t.Fatalf("Decryption key is not of type ECDSA: %#v", key)
	}
	c, _ := middleware.CreateTestContext()
	s, err := CreateApnsJwt(c)
	if err != nil {
		t.Fatalf("CreateApnsJwt failed: %v", err)
	}
	validator := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pubkey, nil
	}
	token, err := jwt.Parse(s, validator, jwt.WithValidMethods([]string{"ES256"}))
	if err != nil {
		t.Fatalf("parse apns jwt failed: %v", err)
	}
	if !token.Valid {
		t.Fatalf("apns jwt is not valid")
	}
	config := storage.GetConfig()
	if token.Header["kid"] != config.ApnsCredId {
		t.Fatalf("token has wrong kid: %q", token.Header["kid"])
	}
	if issuer, err := token.Claims.GetIssuer(); err != nil {
		t.Errorf("get issuer claim failed: %v", err)
	} else if issuer != config.ApnsTeamId {
		t.Errorf("token has the wrong issuer: %q", issuer)
	}
}

func TestValidateClientJwt(t *testing.T) {
	id := uuid.New().String()
	secret := client.MakeNonce()
	key, _ := hex.DecodeString(secret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:   id,
		IssuedAt: jwt.NewNumericDate(time.Now()),
	})
	signed, _ := token.SignedString(key)
	c, _ := middleware.CreateTestContext()
	if !ValidateClientJwt(c, signed, id, secret) {
		t.Errorf("validate client jwt failed")
	}
	if ValidateClientJwt(c, signed, uuid.New().String(), secret) {
		t.Errorf("validated client jwt with wrong client id")
	}
	if ValidateClientJwt(c, signed, id, client.MakeNonce()) {
		t.Errorf("validated client jwt with wrong secret")
	}
	if ValidateClientJwt(c, signed, id, "secret") {
		t.Errorf("validated client jwt with invalid secret")
	}
	old := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:   id,
		IssuedAt: jwt.NewNumericDate(time.Now().Add(-10 * time.Minute)),
	})
	oldSigned, _ := old.SignedString(key)
	if ValidateClientJwt(c, oldSigned, id, secret) {
		t.Errorf("validated client jwt with older issued at")
	}
	newJwt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:   id,
		IssuedAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
	})
	newSigned, _ := newJwt.SignedString(key)
	if ValidateClientJwt(c, newSigned, id, secret) {
		t.Errorf("validated client jwt with newer issued at")
	}
	noIssuer := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ID:       "foo",
		IssuedAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
	})
	noIssuerSigned, _ := noIssuer.SignedString(key)
	if ValidateClientJwt(c, noIssuerSigned, id, secret) {
		t.Errorf("validated client jwt with no issuer")
	}
	noIssuedAt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: id,
		ID:     "foo",
	})
	noIssuedAtSigned, _ := noIssuedAt.SignedString(key)
	if ValidateClientJwt(c, noIssuedAtSigned, id, secret) {
		t.Errorf("validated client jwt with no issued at")
	}
}
