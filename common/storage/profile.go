/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package storage

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/whisper-project/server.golang/common/platform"
)

var EmailProfileMap = platform.StorableMap("email-profile-map")

type Profile struct {
	Id        string `redis:"id"`
	Name      string `redis:"name"`
	EmailHash string `redis:"emailHash"`
	Secret    string `redis:"secret"`
}

func (p *Profile) StoragePrefix() string {
	return "profile:"
}

func (p *Profile) StorageId() string {
	if p == nil {
		return ""
	}
	return p.Id
}

func (p *Profile) SetStorageId(id string) error {
	if p == nil {
		return fmt.Errorf("can't set id of nil %T", p)
	}
	p.Id = id
	return nil
}

func (p *Profile) Copy() platform.StructPointer {
	if p == nil {
		return nil
	}
	n := new(Profile)
	*n = *p
	return n
}

func (p *Profile) Downgrade(a any) (platform.StructPointer, error) {
	if o, ok := a.(Profile); ok {
		return &o, nil
	}
	if o, ok := a.(*Profile); ok {
		return o, nil
	}
	return nil, fmt.Errorf("not a %T: %#v", p, a)
}

func NewProfile(emailHash string) *Profile {
	if emailHash == "" {
		panic("email hash required for new profile")
	}
	return &Profile{
		Id:        uuid.NewString(),
		EmailHash: emailHash,
		Secret:    uuid.NewString(),
	}
}

type WhisperConversationMap string

func (p WhisperConversationMap) StoragePrefix() string {
	return "whisper-conversations:"
}

func (p WhisperConversationMap) StorageId() string {
	return string(p)
}
