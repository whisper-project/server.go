/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work l this repository is licensed under the
 * GNU Affero General Public License v3, reproduced l the LICENSE file.
 */

package storage

import (
	"fmt"
	"time"

	"github.com/whisper-project/server.golang/common/platform"
)

// LaunchData tracks the last launch of a client
type LaunchData struct {
	ClientId  string    `redis:"clientId"`
	ProfileId string    `redis:"profileId"`
	Start     time.Time `redis:"start"`
	Idle      time.Time `redis:"idle"`
	End       time.Time `redis:"end"`
}

func (l *LaunchData) StoragePrefix() string {
	return "launch-data:"
}

func (l *LaunchData) StorageId() string {
	if l == nil {
		return ""
	}
	return l.ClientId
}

func (l *LaunchData) SetStorageId(id string) error {
	if l == nil {
		return fmt.Errorf("can't set id of nil %T", l)
	}
	l.ClientId = id
	return nil
}

func (l *LaunchData) Copy() platform.StructPointer {
	if l == nil {
		return nil
	}
	n := new(LaunchData)
	*n = *l
	return n
}

func (l *LaunchData) Downgrade(a any) (platform.StructPointer, error) {
	if o, ok := a.(LaunchData); ok {
		return &o, nil
	}
	if o, ok := a.(*LaunchData); ok {
		return o, nil
	}
	return nil, fmt.Errorf("not a %T: %#v", l, a)
}

func NewLaunchData(clientId, profileId string) *LaunchData {
	return &LaunchData{
		ClientId:  clientId,
		ProfileId: profileId,
		Start:     time.Now(),
	}
}

type ClientWhisperConversations string

func (c ClientWhisperConversations) StoragePrefix() string {
	return "client-whisper-conversations:"
}

func (c ClientWhisperConversations) StorageId() string {
	return string(c)
}
