/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package profile

import (
	"encoding/json"
)

type Data struct {
	Id                 string   `redis:"id" json:"id"`
	Name               string   `redis:"name" json:"name"`
	Password           string   `redis:"password" json:"password"`
	WhisperTimestamp   string   `redis:"whisperTimestamp" json:"whisperTimestamp"`
	WhisperProfile     string   `redis:"whisperProfile" json:"whisperProfile"`
	ListenTimestamp    string   `redis:"listenTimestamp" json:"listenTimestamp"`
	ListenProfile      string   `redis:"listenProfile" json:"listenProfile"`
	SettingsVersion    int64    `redis:"settingsVersion" json:"settingsVersion"`
	SettingsETag       string   `redis:"settingsETag" json:"settingsETag"`
	SettingsProfile    Settings `redis:"settingsProfile" json:"settingsProfile"`
	FavoritesTimestamp string   `redis:"favoritesTimestamp" json:"favoritesTimestamp"`
	FavoritesProfile   string   `redis:"favoritesProfile" json:"favoritesProfile"`
}

type Settings struct {
	Id       string      `json:"id"`
	Version  float64     `json:"version"`
	Settings AppSettings `json:"settings"`
	ETag     string      `json:"etag"`
}

type AppSettings map[string]string

func (s Settings) MarshalBinary() ([]byte, error) {
	// for saving Settings data to Redis as a JSON blob
	bytes, err := json.Marshal(s.Settings)
	if err != nil {
		return nil, err
	}
	us := uploadedSettings{
		Id:       s.Id,
		Version:  s.Version,
		Settings: string(bytes),
		ETag:     s.ETag,
	}
	return json.Marshal(us)
}

func (s *Settings) UnmarshalText(text []byte) error {
	// for reading Settings data stored to Redis as a JSON blob
	var uploaded uploadedSettings
	if err := json.Unmarshal(text, &uploaded); err != nil {
		return err
	}
	var inner AppSettings
	if err := json.Unmarshal([]byte(uploaded.Settings), &inner); err != nil {
		return err
	}
	*s = Settings{
		Id:       uploaded.Id,
		Version:  uploaded.Version,
		Settings: inner,
		ETag:     uploaded.ETag,
	}
	return nil
}

type uploadedSettings struct {
	Id       string  `json:"id"`
	Version  float64 `json:"version"`
	Settings string  `json:"settings"`
	ETag     string  `json:"etag"`
}

func (p Data) StoragePrefix() string {
	return "pro:"
}

func (p Data) StorageId() string {
	return p.Id
}

type List string

func (pcl List) StoragePrefix() string {
	return "pro-list:"
}

func (pcl List) StorageId() string {
	return string(pcl)
}

type Clients string

func (pcl Clients) StoragePrefix() string {
	return "pro-clients:"
}

func (pcl Clients) StorageId() string {
	return string(pcl)
}
