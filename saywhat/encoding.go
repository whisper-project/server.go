package saywhat

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"clickonetwo.io/whisper/server/profile"
	"clickonetwo.io/whisper/server/storage"
)

type SettingsProfile map[string]interface{}

type WhisperSettings map[string]string

type Settings struct {
	ApiKey             string             `json:"api_key"`
	ApiRoot            string             `json:"api_root"`
	GenerationSettings GenerationSettings `json:"generation_settings"`
}

type GenerationSettings struct {
	OutputFormat             string        `json:"output_format"`
	OptimizeStreamingLatency string        `json:"optimize_streaming_latency"`
	VoiceId                  string        `json:"voice_id"`
	ModelId                  string        `json:"model_id"`
	VoiceSettings            VoiceSettings `json:"voice_settings"`
	PronunciationDictionary  string        `json:"pronunciation_dictionary"`
}

type VoiceSettings struct {
	SimilarityBoost float64 `json:"similarity_boost"`
	Stability       float64 `json:"stability"`
	UseSpeakerBoost bool    `json:"use_speaker_boost"`
}

func defaultSettings() Settings {
	return Settings{
		ApiRoot: "https://api.elevenlabs.io/v1",
		GenerationSettings: GenerationSettings{
			OutputFormat:             "mp3_44100_128",
			OptimizeStreamingLatency: "1",
			VoiceId:                  "pNInz6obpgDQGcFmaJgB", // Adam - free voice
			ModelId:                  "eleven_turbo_v2",
			VoiceSettings: VoiceSettings{
				SimilarityBoost: 0.5,
				Stability:       0.5,
				UseSpeakerBoost: true,
			},
		},
	}
}

func (s *Settings) LoadFromProfile(c *gin.Context, profileId string) error {
	*s = defaultSettings()
	p := &profile.Data{Id: profileId}
	if err := storage.LoadFields(c.Request.Context(), p); err != nil {
		return err
	}
	sp := SettingsProfile{}
	if err := json.Unmarshal([]byte(p.SettingsProfile), &sp); err != nil {
		return err
	}
	if sv, ok := sp["version"].(float64); !ok || sv < 2 {
		// version is too old to have a full set of dictionary settings
		return fmt.Errorf("Profile version %.0f is not supported", sv)
	}
	ws := WhisperSettings{}
	if err := json.Unmarshal([]byte(p.SettingsProfile), &ws); err != nil {
		return err
	}
	if i, err := strconv.Atoi(ws["elevenlabs_latency_reduction_preference"]); err == nil {
		return err
	} else {
		s.GenerationSettings.OptimizeStreamingLatency = strconv.Itoa(i + 1)
	}
	s.ApiKey = ws["elevenlabs_api_key_preference"]
	s.GenerationSettings.VoiceId = ws["elevenlabs_voice_id_preference"]
	id1 := ws["elevenlabs_dictionary_id_preference"]
	id2 := ws["elevenlabs_dictionary_version_preference"]
	if id1 != "" && id2 != "" {
		s.GenerationSettings.PronunciationDictionary = fmt.Sprintf("%s|%s", id1, id2)
	}
	return nil
}

func (s *Settings) StoreToProfile(c *gin.Context, profileId string) error {
	p := &profile.Data{Id: profileId}
	if err := storage.LoadFields(c.Request.Context(), p); err != nil {
		return err
	}
	sp := SettingsProfile{}
	if err := json.Unmarshal([]byte(p.SettingsProfile), &sp); err != nil {
		return err
	}
	if sv, ok := sp["version"].(float64); !ok || sv < 2 {
		// version is too old to have a full set of dictionary settings
		return fmt.Errorf("settings profile version (%.0f) too old to set")
	}
	ws := WhisperSettings{}
	if err := json.Unmarshal([]byte(p.SettingsProfile), &ws); err != nil {
		return err
	}
	ws["elevenlabs_api_key_preference"] = s.ApiKey
	ws["elevenlabs_voice_id_preference"] = s.GenerationSettings.VoiceId
	if i, err := strconv.Atoi(s.GenerationSettings.OptimizeStreamingLatency); err == nil {
		return err
	} else {
		ws["elevenlabs_latency_reduction_preference"] = strconv.Itoa(i + 1)
	}
	if s.GenerationSettings.PronunciationDictionary == "" {
		ws["elevenlabs_dictionary_id_preference"] = ""
		ws["elevenlabs_dictionary_version_preference"] = ""
	} else {
		ids := strings.Split(s.GenerationSettings.PronunciationDictionary, "|")
		if len(ids) != 2 {
			return fmt.Errorf("invalid pronunciation dictionary locator: %s", s.GenerationSettings.PronunciationDictionary)
		}
		ws["elevenlabs_dictionary_id_preference"] = ids[0]
		ws["elevenlabs_dictionary_version_preference"] = ids[1]
	}
	js, err := json.Marshal(ws)
	if err != nil {
		return err
	}
	eTag := fmt.Sprintf("%02x", md5.Sum(js))
	sp["settings"] = string(js)
	sp["eTag"] = eTag
	js, err = json.Marshal(sp)
	if err != nil {
		return err
	}
	p.SettingsProfile = string(js)
	p.SettingsETag = eTag
	return storage.SaveFields(c.Request.Context(), p)
}
