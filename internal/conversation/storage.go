/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package conversation

import (
	"fmt"

	"clickonetwo.io/whisper/internal/storage"
)

type Data struct {
	Id      string `redis:"id"`
	Name    string `redis:"name"`
	OwnerId string `redis:"ownerId"`
	StateId string `redis:"stateId"`
}

func (c *Data) StoragePrefix() string {
	return "con:"
}

func (c *Data) StorageId() string {
	if c == nil {
		return ""
	}
	return c.Id
}

func (c *Data) SetStorageId(id string) error {
	if c == nil {
		return fmt.Errorf("can't set storage id of nil struct")
	}
	c.Id = id
	return nil
}

func (c *Data) Copy() any {
	if c == nil {
		return nil
	}
	n := new(Data)
	*n = *c
	return any(n)
}

func (c Data) Downgrade(in any) (storage.StorableStruct, error) {
	if o, ok := in.(Data); ok {
		return &o, nil
	}
	return nil, fmt.Errorf("not a conversation.Data: %#v", in)
}

type State struct {
	Id             string `redis:"id"`
	ServerId       string `redis:"serverId"`
	ConversationId string `redis:"conversationId"`
	ContentId      string `redis:"contentId"`
	TzId           string `redis:"tzId"`
	StartTime      int64  `redis:"startTime"`
	Duration       int64  `redis:"duration"`
	ContentKey     string `redis:"contentKey"`
	Transcription  string `redis:"transcription"`
	ErrCount       int64  `redis:"errCount"`
	Ttl            int64  `redis:"ttl"`
}

func (s *State) StoragePrefix() string {
	return "tra:"
}

func (s *State) StorageId() string {
	if s == nil {
		return ""
	}
	return s.Id
}

func (s *State) SetStorageId(id string) error {
	if s == nil {
		return fmt.Errorf("can't set storage id of nil struct")
	}
	s.Id = id
	return nil
}

func (s *State) Copy() any {
	if s == nil {
		return nil
	}
	n := new(State)
	*n = *s
	return any(n)
}

func (s State) Downgrade(in any) (storage.StorableStruct, error) {
	if o, ok := in.(State); ok {
		return &o, nil
	}
	return nil, fmt.Errorf("not a conversation.State: %#v", in)
}
