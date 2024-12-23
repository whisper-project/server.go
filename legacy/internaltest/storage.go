/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package internaltest

import (
	"context"

	storage2 "github.com/whisper-project/server.golang/common/storage"

	"github.com/google/uuid"
)

//goland:noinspection SpellCheckingInspection
var (
	KnownUserId           = "B11C1B3D-21E6-4766-B16B-4FDEED785139"
	KnownUserName         = "Dan Brotsky"
	KnownClientId         = "561E5E8E-EA35-405A-A256-69C74713FAFD"
	KnownClientUserName   = "Dan Brotsky"
	KnownConversationId   = "3C6CE484-4A73-4D06-A8B9-4FC8EF51F5BA"
	KnownConversationName = "Anyone"
	KnownStateId          = "d7dfb2b5-f25a-4de7-8c4a-52af08f1e7f3"
)

func NewTestId() string {
	return "test-" + uuid.NewString()
}

func RemoveCreatedTestData() {
	ctx := context.Background()
	db, prefix := storage2.GetDb()
	iter := db.Scan(ctx, 0, prefix+"*:test-*", 20).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		_ = db.Del(ctx, key)
	}
}

func LoadAndCopy(o storage2.StructPointer) (storage2.StructPointer, error) {
	if err := storage2.LoadFields(context.Background(), o); err != nil {
		return o, err
	}
	c := o.Copy()
	if err := c.SetStorageId(NewTestId()); err != nil {
		return o, err
	}
	return c, nil
}

func LoadCopyAndSave(o storage2.StructPointer) (storage2.StructPointer, error) {
	c, err := LoadAndCopy(o)
	if err != nil {
		return o, err
	}
	if err := storage2.SaveFields(context.Background(), c); err != nil {
		return o, err
	}
	return c, nil
}
