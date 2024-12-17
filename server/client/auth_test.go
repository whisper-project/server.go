/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package client

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"clickonetwo.io/whisper/server/middleware"
	"clickonetwo.io/whisper/server/storage"
)

func TestHasAuthChanged(t *testing.T) {
	c, _ := middleware.CreateTestContext()
	received := Data{Id: uuid.New().String(), ProfileId: uuid.New().String()}
	yes, why := received.HasAuthChanged(c)
	if yes != true || why != "APNS token from new" {
		t.Errorf("HasAuthChanged() = %v, %q; want %v, %q", yes, why, true, "APNS token from new")
	}
	if err := storage.SaveFields(c.Request.Context(), &received); err != nil {
		t.Errorf("Failed to save stored data for client %q: %v", received.Id, err)
	}
	yes, why = received.HasAuthChanged(c)
	if yes != false || why != "" {
		t.Errorf("HasAuthChanged() = %v, %q; want %v, %q", yes, why, false, "")
	}
	received.LastSecret = "secret"
	yes, why = received.HasAuthChanged(c)
	if yes != true || why != "unconfirmed secret from existing" {
		t.Errorf("HasAuthChanged() = %v, %q; want %v, %q", yes, why, true, "unconfirmed secret from existing")
	}
	if err := storage.SaveFields(c.Request.Context(), &received); err != nil {
		t.Errorf("Failed to save stored data for client %q: %v", received.Id, err)
	}
	received.Token = "token"
	yes, why = received.HasAuthChanged(c)
	if yes != true || why != "new APNS token from existing" {
		t.Errorf("HasAuthChanged() = %v, %q; want %v, %q", yes, why, true, "new APNS token from existing")
	}
	if err := storage.SaveFields(c.Request.Context(), &received); err != nil {
		t.Errorf("Failed to save stored data for client %q: %v", received.Id, err)
	}
	received.AppInfo = "app info"
	yes, why = received.HasAuthChanged(c)
	if yes != true || why != "new build data from existing" {
		t.Errorf("HasAuthChanged() = %v, %q; want %v, %q", yes, why, true, "new build data from existing")
	}
	if err := storage.SaveFields(c.Request.Context(), &received); err != nil {
		t.Errorf("Failed to save stored data for client %q: %v", received.Id, err)
	}
	if err := storage.DeleteStorage(c.Request.Context(), &received); err != nil {
		t.Errorf("Failed to delete stored data for client %q: %v", received.Id, err)
	}
}

func TestRefreshSecret(t *testing.T) {
	c, _ := middleware.CreateTestContext()
	received := Data{Id: uuid.New().String()}
	if _, err := received.RefreshSecret(c, false); err == nil {
		t.Errorf("RefreshSecret with no token got no error")
	}
	received.Token = "token"
	if refreshed, err := received.RefreshSecret(c, false); err != nil {
		t.Errorf("RefreshSecret with token got error: %v", err)
	} else if !refreshed {
		t.Errorf("RefreshSecret with token but no secret didn't get refreshed")
	}
	secret1 := received.Secret
	date1 := received.SecretDate
	if secret1 == "" {
		t.Errorf("After first refresh: secret is empty")
	}
	if date1 != 0 {
		t.Errorf("After first refresh: date is not 0: %d", date1)
	}
	if refreshed, err := received.RefreshSecret(c, false); err != nil {
		t.Errorf("RefreshSecret with token and secret but no date got error: %v", err)
	} else if !refreshed {
		t.Errorf("RefreshSecret with token and secret but no date didn't get refreshed")
	}
	secret2 := received.Secret
	date2 := received.SecretDate
	if secret2 != secret1 {
		t.Errorf("After second refresh: secret differs: was %q, is now %q", secret1, secret2)
	}
	if date2 != date1 {
		t.Errorf("After second refresh: date differs: was %d, is now %d", date1, date2)
	}
	received.SecretDate = time.Now().UnixMilli()
	if refreshed, err := received.RefreshSecret(c, false); err != nil {
		t.Errorf("RefreshSecret with token and secret and date got error: %v", err)
	} else if refreshed {
		t.Errorf("RefreshSecret with token and secret and date got refreshed")
	}
	if refreshed, err := received.RefreshSecret(c, true); err != nil {
		t.Errorf("RefreshSecret forced with token and secret and date got error: %v", err)
	} else if !refreshed {
		t.Errorf("RefreshSecret forced with token and secret and date didn't get refreshed")
	}
	secret3 := received.Secret
	date3 := received.SecretDate
	if secret3 == "" || secret3 == secret2 {
		t.Errorf("After forced refresh, secret is wrong: was %q, is now %q", secret2, secret3)
	}
	if date3 != 0 {
		t.Errorf("After forced refresh: date is not 0: %d", date1)
	}
	if refreshed, err := received.RefreshSecret(c, true); err != nil {
		t.Errorf("RefreshSecret forced with token and secret and no date got error: %v", err)
	} else if !refreshed {
		t.Errorf("RefreshSecret forced with token and secret and no date didn't get refreshed")
	}
	secret4 := received.Secret
	date4 := received.SecretDate
	if secret4 != secret3 {
		t.Errorf("After second refresh: secret differs: was %q, is now %q", secret1, secret2)
	}
	if date4 != date3 {
		t.Errorf("After second refresh: date differs: was %d, is now %d", date1, date2)
	}
	if err := storage.DeleteStorage(c.Request.Context(), &received); err != nil {
		t.Errorf("Failed to delete stored data for client %q: %v", received.Id, err)
	}
}
