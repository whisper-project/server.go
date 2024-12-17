/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package saywhat

import (
	"testing"

	"clickonetwo.io/whisper/server/internaltest"
	"clickonetwo.io/whisper/server/middleware"
)

func TestSettings_AddMissingSettings(t *testing.T) {
	var s Settings
	s.GenerationSettings.VoiceSettings.Stability = 0.8
	s.GenerationSettings.VoiceId = "test voice id"
	s.addMissingSettings()
	if s.ApiRoot != "https://api.elevenlabs.io/v1" {
		t.Errorf("ApiRoot is wrong")
	}
	if s.ApiKey != "" {
		t.Errorf("ApiKey is wrong")
	}
	if s.GenerationSettings.OutputFormat != "mp3_44100_128" {
		t.Errorf("OutputFormat is wrong")
	}
	if s.GenerationSettings.OptimizeStreamingLatency != "1" {
		t.Errorf("OptimizeStreamingLatency is wrong")
	}
	if s.GenerationSettings.VoiceId != "test voice id" {
		t.Errorf("VoiceId is wrong")
	}
	if s.GenerationSettings.ModelId != "eleven_turbo_v2" {
		t.Errorf("ModelId is wrong")
	}
	if s.GenerationSettings.VoiceSettings.SimilarityBoost != 0.5 {
		t.Errorf("SimilarityBoost is wrong")
	}
	if s.GenerationSettings.VoiceSettings.Stability != 0.8 {
		t.Errorf("Stability is wrong")
	}
	if s.GenerationSettings.VoiceSettings.UseSpeakerBoost != true {
		t.Errorf("UseSpeakerBoost is wrong")
	}
}

func TestSettings_LoadFromProfile(t *testing.T) {
	var s Settings
	c, _ := middleware.CreateTestContext()
	if err := s.LoadFromProfile(c, internaltest.KnownUserId); err != nil {
		t.Fatal(err)
	}
}
