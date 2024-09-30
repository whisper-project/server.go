package client

import (
	"context"
	"errors"
	"os"
	"testing"

	storage2 "clickonetwo.io/whisper/server/internal/storage"
)

func TestCountLegacyClients(t *testing.T) {
	if os.Getenv("DO_LEGACY_TESTS") != "YES" {
		t.Skip("Skipping legacy client test")
	}
	if err := storage2.PushConfig("../.env.production"); err != nil {
		t.Fatalf("Can't load production config: %v", err)
	}
	defer storage2.PopConfig()
	data := Data{}
	count := 0
	countOnly := func() {
		count++
	}
	ctx := context.Background()
	if err := storage2.MapFields(ctx, countOnly, &data); err != nil {
		if errors.Is(err, ctx.Err()) {
			t.Logf("Found %d production clients before timing out.", count)
		} else {
			t.Errorf("Failed to map production data: %v", err)
		}
	} else {
		t.Logf("Found %d clients in production.", count)
	}
}
