/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package client

import (
	"context"
	"errors"
	"os"
	"testing"

	"clickonetwo.io/whisper/internal/storage"
)

func TestCountLegacyClients(t *testing.T) {
	if os.Getenv("DO_LEGACY_TESTS") != "YES" {
		t.Skip("Skipping legacy client test")
	}
	if err := storage.PushConfig("production"); err != nil {
		t.Fatalf("Can't load production config: %v", err)
	}
	defer storage.PopConfig()
	data := Data{}
	count := 0
	countOnly := func() {
		count++
	}
	ctx := context.Background()
	if err := storage.MapFields(ctx, countOnly, &data); err != nil {
		if errors.Is(err, ctx.Err()) {
			t.Logf("Found %d production clients before timing out.", count)
		} else {
			t.Errorf("Failed to map production data: %v", err)
		}
	} else {
		t.Logf("Found %d clients in production.", count)
	}
}
