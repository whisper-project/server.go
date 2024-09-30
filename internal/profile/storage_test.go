package profile

import (
	"context"
	"os"
	"testing"

	storage2 "clickonetwo.io/whisper/server/internal/storage"
)

func TestEnumerateLegacyProfiles(t *testing.T) {
	if os.Getenv("DO_LEGACY_TESTS") != "YES" {
		t.Skip("Skipping legacy encoding test")
	}
	if err := storage2.PushConfig("../.env.production"); err != nil {
		t.Fatalf("Can't load production config: %v", err)
	}
	defer storage2.PopConfig()
	d := &Data{}
	total := 0
	named := 0
	settings := 0
	report := func() {
		total++
		if d.Name != "" {
			named++
		}
		if len(d.SettingsProfile) > 0 {
			settings++
		}
	}
	if err := storage2.MapFields(context.Background(), report, d); err != nil {
		t.Fatal(err)
	}
	t.Logf("Found %d shared profiles (%d named) of which %d had settings profiles", total, named, settings)
}
