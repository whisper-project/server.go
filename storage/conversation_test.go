package storage

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestCountLegacyConversations(t *testing.T) {
	if os.Getenv("DO_LEGACY_TESTS") != "YES" {
		t.Skip("Skipping legacy client test")
	}
	if err := PushConfig("../.env.production"); err != nil {
		t.Fatalf("Can't load production config: %v", err)
	}
	defer PopConfig()
	data := ConversationState{}
	count := 0
	earliest := time.Now()
	latest := time.UnixMilli(0)
	doCount := func() {
		count++
		start := time.UnixMilli(data.StartTime)
		if start.After(latest) {
			latest = start
		}
		if start.Before(earliest) {
			earliest = start
		}
	}
	ctx := context.Background()
	if err := MapFields(ctx, doCount, &data); err != nil {
		t.Errorf("Failed to map production data: %v", err)
	} else {
		t.Logf("Found %d transcripts, earliest at %v, latest at %v", count, earliest, latest)
	}
}
