/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package internaltest

import (
	"context"
	"testing"

	"github.com/go-test/deep"

	"github.com/whisper-project/server.golang/internal/profile"
	"github.com/whisper-project/server.golang/internal/storage"
)

func TestLoadAndCopy(t *testing.T) {
	var s1, s2 storage.StructPointer
	var err error
	s1 = &profile.UserProfile{Id: KnownUserId}
	s2, err = LoadAndCopy(s1)
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(s1, s2); len(diff) != 1 || diff[0][0:3] != "Id:" {
		t.Error(diff)
	}
}

func TestLoadCopyAndSave(t *testing.T) {
	var s1, s2, s3 storage.StructPointer
	var err error
	s1 = &profile.UserProfile{Id: KnownUserId}
	if s2, err = LoadCopyAndSave(s1); err != nil {
		t.Fatal(err)
	}
	s3 = &profile.UserProfile{Id: s2.StorageId()}
	if err = storage.LoadFields(context.Background(), s3); err != nil {
		t.Error(err)
	}
	if diff := deep.Equal(s1, s3); len(diff) != 1 || diff[0][0:3] != "Id:" {
		t.Error(diff)
	}
}

func TestRemoveCreatedTestData(t *testing.T) {
	s1 := &profile.UserProfile{Id: KnownUserId}
	s2, err := LoadCopyAndSave(s1)
	if err != nil {
		t.Fatal(err)
	}
	RemoveCreatedTestData()
	s3 := &profile.UserProfile{Id: s2.StorageId()}
	if err = storage.LoadFields(context.Background(), s3); err == nil {
		t.Errorf("Load should fail")
	}
}
