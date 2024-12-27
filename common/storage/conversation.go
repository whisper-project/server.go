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

type Conversation struct {
	Id    string `redis:"id"`
	Owner string `redis:"owner"`
	Name  string `redis:"name"`
}

func (c *Conversation) StoragePrefix() string {
	return "conversation:"
}

func (c *Conversation) StorageId() string {
	if c == nil {
		return ""
	}
	return c.Id
}

func (c *Conversation) SetStorageId(id string) error {
	if c == nil {
		return fmt.Errorf("can't set id of nil %T", c)
	}
	c.Id = id
	return nil
}

func (c *Conversation) Copy() platform.StructPointer {
	if c == nil {
		return nil
	}
	n := new(Conversation)
	*n = *c
	return n
}

func (c *Conversation) Downgrade(a any) (platform.StructPointer, error) {
	if o, ok := a.(Conversation); ok {
		return &o, nil
	}
	if o, ok := a.(*Conversation); ok {
		return o, nil
	}
	return nil, fmt.Errorf("not a %T: %#v", c, a)
}

func NewConversation(owner, name string) *Conversation {
	return &Conversation{
		Id:    uuid.NewString(),
		Owner: owner,
		Name:  name,
	}
}

type AllowedParticipantSet string

func (a AllowedParticipantSet) StoragePrefix() string {
	return "allowed-participants"
}

func (a AllowedParticipantSet) StorageId() string {
	return string(a)
}
