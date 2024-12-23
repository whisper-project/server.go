/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package profile

import (
	"fmt"

	"github.com/whisper-project/server.golang/common/storage"
)

type StoredMap string

func (s StoredMap) StoragePrefix() string {
	return "map:"
}

func (s StoredMap) StorageId() string {
	return string(s)
}

var (
	EmailProfileMap  = StoredMap("email_profile_map")
	ClientProfileMap = StoredMap("client_profile_map")
)

type Profile struct {
	Id        string `redis:"id"`
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
		return fmt.Errorf("can't set storage id of nil struct")
	}
	p.Id = id
	return nil
}

func (p *Profile) Copy() storage.StructPointer {
	if p == nil {
		return nil
	}
	n := new(Profile)
	*n = *p
	return n
}

func (p *Profile) Downgrade(in any) (storage.StructPointer, error) {
	if o, ok := in.(Profile); ok {
		return &o, nil
	}
	if o, ok := in.(*Profile); ok {
		return o, nil
	}
	return nil, fmt.Errorf("not a %T: %#v", p, in)
}

type WhisperConversationMap string

func (p WhisperConversationMap) StoragePrefix() string {
	return "whisper-conversations:"
}

func (p WhisperConversationMap) StorageId() string {
	return string(p)
}
