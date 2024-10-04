/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package profile

import (
	"context"
	"os"
	"testing"

	"clickonetwo.io/whisper/server/internal/storage"
)

func TestEnumerateProfiles(t *testing.T) {
	d := &Data{}
	total := 0
	named := 0
	settings := 0
	report := func() {
		total++
		if d.Name != "" {
			named++
		}
		if d.SettingsProfile.Settings["elevenlabs_api_key_preference"] != "" {
			settings++
		}
	}
	if err := storage.MapFields(context.Background(), report, d); err != nil {
		t.Fatal(err)
	}
	t.Logf("Found %d shared profiles (%d named) of which %d had elevenlabs keys", total, named, settings)
}

func TestEnumerateLegacyProfiles(t *testing.T) {
	if os.Getenv("DO_LEGACY_TESTS") != "YES" {
		t.Skip("Skipping legacy encoding test")
	}
	if err := storage.PushConfig("production"); err != nil {
		t.Fatalf("Can't load production config: %v", err)
	}
	defer storage.PopConfig()
	d := &Data{}
	total := 0
	named := 0
	settings := 0
	report := func() {
		total++
		if d.Name != "" {
			named++
		}
		if d.SettingsProfile.Settings["elevenlabs_api_key_preference"] != "" {
			settings++
		}
	}
	if err := storage.MapFields(context.Background(), report, d); err != nil {
		t.Fatal(err)
	}
	t.Logf("Found %d shared profiles (%d named) of which %d had elevenlabs keys", total, named, settings)
}
