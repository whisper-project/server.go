package saywhat

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"clickonetwo.io/whisper/server/profile"
	"clickonetwo.io/whisper/server/storage"
)

func TestEnumerateLegacySettingsProfiles(t *testing.T) {
	if os.Getenv("DO_LEGACY_TESTS") != "YES" {
		t.Skip("Skipping legacy encoding test")
	}
	if err := storage.PushConfig("../.env.production"); err != nil {
		t.Fatalf("Can't load production config: %v", err)
	}
	defer storage.PopConfig()
	d := &profile.Data{}
	count := 0
	report := func() {
		if d.SettingsProfile != "" {
			sp := make(map[string]interface{})
			if err := json.Unmarshal([]byte(d.SettingsProfile), &sp); err != nil {
				t.Fatalf("Can't unmarshal settings profile (%s): %v", d.SettingsProfile, err)
			}
			if v, ok := sp["version"].(float64); !ok || v < 2 {
				// too old to have a pronunciation dictionary
				return
			}
			ws := WhisperSettings{}
			if err := json.Unmarshal([]byte(sp["settings"].(string)), &ws); err != nil {
				t.Fatalf("Can't unmarshal settings part of settings profile (%s): %v", sp["settings"], err)
			}
			if ws["elevenlabs_api_key_preference"] != "" {
				count++
				if ws["elevenlabs_dictionary_id_preference"] != "" {
					fmt.Printf("Profile %s has eTag %s and a dictionary: %v\n", d.Id, d.SettingsETag, sp["settings"])
				}
			}
		}
	}
	if err := storage.MapFields(context.Background(), report, d); err != nil {
		t.Fatal(err)
	}
	t.Logf("Found %d elevenlabs settings profiles", count)
}
