/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package storage

import (
	"testing"
	"time"

	"github.com/whisper-project/server.golang/common/platform"

	"github.com/google/uuid"
)

func TestLaunchDataInterface(t *testing.T) {
	clientId := uuid.NewString()
	l := &LaunchData{ClientId: clientId}
	var n *LaunchData
	platform.StorableInterfaceTester(t, l, "launch-data:", clientId)
	platform.StructPointerInterfaceTester(t, n, l, *l, "launch-data:", clientId)
}

func TestNewLaunchData(t *testing.T) {
	clientId := uuid.NewString()
	profileId := uuid.NewString()
	now := time.Now()
	l := NewLaunchData(clientId, profileId)
	if l.ClientId != clientId {
		t.Errorf("NewLaunchData returned wrong client id. Got %s, Want %s", l.ClientId, clientId)
	}
	if l.ProfileId != profileId {
		t.Errorf("NewLaunchData returned wrong profile id. Got %s, Want %s", l.ProfileId, profileId)
	}
	if !now.Before(l.Start) {
		t.Errorf("NewLaunchData returned an early start. Got %v, Want later than %v", l.Start, now)
	}
	if !now.After(l.Idle) {
		t.Errorf("NewLaunchData returned a late idle. Got %v, Want later than %v", l.Idle, now)
	}
	if !now.After(l.End) {
		t.Errorf("NewLaunchData returned a late end. Got %v, Want later than %v", l.End, now)
	}
}

func TestClientWhisperConversationsInterface(t *testing.T) {
	clientId := uuid.NewString()
	platform.StorableInterfaceTester(t, ClientWhisperConversations(clientId), "client-whisper-conversations:", clientId)
}
