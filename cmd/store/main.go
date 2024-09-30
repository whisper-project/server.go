package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"clickonetwo.io/whisper/server/api/saywhat"
	"clickonetwo.io/whisper/server/internal/profile"
	"clickonetwo.io/whisper/server/internal/storage"
)

type ProfileList string

func (pcl ProfileList) StoragePrefix() string {
	return "test-profiles:"
}

func (pcl ProfileList) StorageId() string {
	return string(pcl)
}

func main() {
	var ps []profile.Data
	p := &profile.Data{}
	pushProfile := func() {
		if len(p.SettingsProfile) > 0 {
			sp := saywhat.SettingsProfile{}
			if err := json.Unmarshal([]byte(p.SettingsProfile), &sp); err != nil {
				panic(err)
			}
			if sp.Version < 4 {
				// too old to have a pronunciation dictionary
				return
			}
			ws := saywhat.WhisperSettings{}
			if err := json.Unmarshal([]byte(sp.Settings), &ws); err != nil {
				panic(err)
			}
			if ws["elevenlabs_api_key_preference"] != "" && ws["elevenlabs_dictionary_id_preference"] != "" {
				ps = append(ps, *p)
			}
		}
	}

	if err := storage.PushConfig(".env.production"); err != nil {
		panic(err)
	}
	if err := storage.MapFields(context.Background(), pushProfile, p); err != nil {
		panic(err)
	}
	fmt.Printf("Found %+d profiles to transfer...\n", len(ps))
	storage.PopConfig()

	var testList ProfileList = "test-profile-list"
	var ids []string
	for i, tp := range ps {
		tp.Id = "test-profile-" + strconv.Itoa(i)
		if err := storage.SaveFields(context.Background(), &tp); err != nil {
			panic(err)
		}
		ids = append(ids, tp.Id)
	}
	if err := storage.PushRange(context.Background(), testList, false, ids...); err != nil {
		panic(err)
	}
	fmt.Printf("Transferred profile IDs are in %q with prefix %q", testList.StorageId(), testList.StoragePrefix())
}
