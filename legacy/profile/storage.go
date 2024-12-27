/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package profile

import (
	"encoding/json"
	"fmt"

	"github.com/whisper-project/server.golang/common/platform"
)

type UserProfile struct {
	Id                 string         `redis:"id" json:"id"`
	LastUsed           int64          `redis:"lastUsed" json:"lastUsed"`
	Name               string         `redis:"name" json:"name"`
	Password           string         `redis:"password" json:"password"`
	WhisperTimestamp   string         `redis:"whisperTimestamp" json:"whisperTimestamp"`
	WhisperProfile     WhisperProfile `redis:"whisperProfile" json:"whisperProfile"`
	ListenTimestamp    string         `redis:"listenTimestamp" json:"listenTimestamp"`
	ListenProfile      ListenProfile  `redis:"listenProfile" json:"listenProfile"`
	SettingsVersion    int64          `redis:"settingsVersion" json:"settingsVersion"`
	SettingsETag       string         `redis:"settingsETag" json:"settingsETag"`
	SettingsProfile    AppSettings    `redis:"settingsProfile" json:"settingsProfile"`
	FavoritesTimestamp string         `redis:"favoritesTimestamp" json:"favoritesTimestamp"`
	FavoritesProfile   AppFavorites   `redis:"favoritesProfile" json:"favoritesProfile"`
}

func (p *UserProfile) StoragePrefix() string {
	return "pro:"
}

func (p *UserProfile) StorageId() string {
	if p == nil {
		return ""
	}
	return p.Id
}

func (p *UserProfile) SetStorageId(id string) error {
	if p == nil {
		return fmt.Errorf("can't set platform id of nil struct")
	}
	p.Id = id
	return nil
}

func (p *UserProfile) Copy() platform.StructPointer {
	if p == nil {
		return nil
	}
	n := new(UserProfile)
	*n = *p
	return n
}

func (p *UserProfile) Downgrade(in any) (platform.StructPointer, error) {
	if o, ok := in.(UserProfile); ok {
		return &o, nil
	}
	if o, ok := in.(*UserProfile); ok {
		return o, nil
	}
	return nil, fmt.Errorf("not a profile.UserProfile: %#v", in)
}

type WhisperProfile struct {
	Id        string                   `json:"id"`
	Timestamp int64                    `json:"timestamp"`
	DefaultId string                   `json:"defaultId"`
	LastId    string                   `json:"lastId"`
	Table     map[string]WhisperRecord `json:"table"`
}

type WhisperRecord struct {
	Id      string            `json:"id"`
	Name    string            `json:"name"`
	Allowed map[string]string `json:"allowed"`
}

func (w WhisperProfile) MarshalBinary() ([]byte, error) {
	return json.Marshal(UploadedWhisperProfile(w))
}

func (w *WhisperProfile) UnmarshalText(data []byte) error {
	var uw UploadedWhisperProfile
	if err := json.Unmarshal(data, &uw); err != nil {
		return err
	}
	*w = WhisperProfile(uw)
	return nil
}

func (w *WhisperProfile) UnmarshalJSON(data []byte) error {
	return w.UnmarshalText(data)
}

type ListenProfile struct {
	Id        string                  `json:"id"`
	Timestamp int64                   `json:"timestamp"`
	Table     map[string]ListenRecord `json:"table"`
}

type ListenRecord struct {
	Id           string  `json:"id"`
	Name         string  `json:"name"`
	Owner        string  `json:"owner"`
	OwnerName    string  `json:"ownerName"`
	LastListened float64 `json:"lastListened"`
}

func (l ListenProfile) MarshalBinary() ([]byte, error) {
	return json.Marshal(UploadedListenProfile(l))
}

func (l *ListenProfile) UnmarshalText(data []byte) error {
	var ul UploadedListenProfile
	if err := json.Unmarshal(data, &ul); err != nil {
		return err
	}
	*l = ListenProfile(ul)
	return nil
}

func (l *ListenProfile) UnmarshalJSON(data []byte) error {
	return l.UnmarshalText(data)
}

type AppSettings struct {
	Id       string           `json:"id"`
	Version  int64            `json:"version"`
	Settings AppInnerSettings `json:"settings"`
	ETag     string           `json:"eTag"`
}

type AppInnerSettings map[string]string

func (s AppSettings) MarshalBinary() ([]byte, error) {
	// for saving AppSettings data to Redis as a JSON blob
	bytes, err := json.Marshal(s.Settings)
	if err != nil {
		return nil, err
	}
	us := UploadedAppSettings{
		Id:       s.Id,
		Version:  s.Version,
		Settings: string(bytes),
		ETag:     s.ETag,
	}
	return json.Marshal(us)
}

func (s *AppSettings) UnmarshalText(text []byte) error {
	// for reading AppSettings data stored to Redis as a JSON blob
	var uploaded UploadedAppSettings
	if err := json.Unmarshal(text, &uploaded); err != nil {
		return err
	}
	var inner AppInnerSettings
	if err := json.Unmarshal([]byte(uploaded.Settings), &inner); err != nil {
		return err
	}
	*s = AppSettings{
		Id:       uploaded.Id,
		Version:  uploaded.Version,
		Settings: inner,
		ETag:     uploaded.ETag,
	}
	return nil
}

type ExpandedAppSettings AppSettings

func (s *AppSettings) UnmarshalJSON(data []byte) error {
	var e ExpandedAppSettings
	if err := json.Unmarshal(data, &e); err != nil {
		return err
	}
	*s = AppSettings(e)
	return nil
}

type AppFavorites struct {
	Id         string              `json:"id"`
	Timestamp  int64               `json:"timestamp"`
	Favorites  []AppFavorite       `json:"favorites"`
	GroupList  []string            `json:"groupList"`
	GroupTable map[string][]string `json:"groupTable"`
}

type AppFavorite struct {
	Name       string `json:"name"`
	Text       string `json:"text"`
	SpeechId   string `json:"speechId,omitempty"`
	SpeechHash string `json:"speechHash,omitempty"`
}

func (f AppFavorites) MarshalBinary() ([]byte, error) {
	return json.Marshal(UploadedAppFavorites(f))
}

func (f *AppFavorites) UnmarshalText(text []byte) error {
	var af UploadedAppFavorites
	if err := json.Unmarshal(text, &af); err != nil {
		return err
	}
	*f = AppFavorites(af)
	return nil
}

func (f *AppFavorites) UnmarshalJSON(data []byte) error {
	return f.UnmarshalText(data)
}
