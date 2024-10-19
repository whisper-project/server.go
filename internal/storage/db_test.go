/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package storage

import (
	"testing"
)

func TestGetDb(t *testing.T) {
	db, prefix := GetDb()
	if db == nil || prefix != "t:" {
		t.Errorf("initial GetDb didn't return test db: %v, %q", db, prefix)
	}
}

func TestGetMultiDifferentDbs(t *testing.T) {
	dbT, prefixT := GetDb()
	t.Logf("Initial test database is: %v, %q", dbT, prefixT)
	if err := PushConfig("development"); err != nil {
		t.Fatalf("failed to push development config: %v", err)
	}
	dbD, prefixD := GetDb()
	if dbT == dbD || prefixT == prefixD {
		t.Fatalf("Dbs before and after dev push are the same: %v & %v, %q & %q", dbT, dbD, prefixT, prefixD)
	}
	if dbD == nil || prefixD != "d:" {
		t.Fatalf("GetDb didn't return dev db after push: %v, %q", dbD, prefixD)
	} else {
		t.Logf("Pushed dev database is: %v, %q", dbD, prefixD)
	}
	if err := PushConfig("staging"); err != nil {
		t.Fatalf("failed to push staging config: %v", err)
	}
	dbS, prefixS := GetDb()
	if dbD == dbS || prefixD == prefixS {
		t.Fatalf("Dbs before and after stage push are the same: %v & %v, %q & %q", dbD, dbS, prefixD, prefixS)
	}
	if dbS == nil || prefixS != "s:" {
		t.Fatalf("GetDb didn't return staging db after push: %v, %q", dbS, prefixS)
	} else {
		t.Logf("Pushed staging database is: %v, %q", dbS, prefixS)
	}
	PopConfig()
	dbD2, prefixD2 := GetDb()
	if prefixD2 != prefixD {
		t.Fatalf("Dev prefix after pop is %q", prefixD2)
	}
	if dbD2 == dbD {
		t.Errorf("Dev db before and after pop are the same: %v", dbD)
	}
	if err := PushConfig("production"); err != nil {
		t.Fatalf("failed to push production config: %v", err)
	}
	dbP, prefixP := GetDb()
	if dbP == dbD2 || prefixP == prefixD2 {
		t.Fatalf("Dbs before and after prod push are the same: %v & %v, %q & %q", dbP, dbD2, prefixP, prefixD2)
	}
	if dbP == nil || prefixP != "p:" {
		t.Fatalf("GetDb didn't return prod db after push: %v, %q", dbP, prefixP)
	} else {
		t.Logf("Pushed prod database is: %v, %q", dbP, prefixP)
	}
	PopConfig()
	dbD3, prefixD3 := GetDb()
	if prefixD3 != prefixD2 {
		t.Fatalf("Dev prefix after pop is %q", prefixD3)
	}
	if dbD3 == dbD2 {
		t.Errorf("Dev db before and after pop are the same: %v", dbD3)
	}
	PopConfig()
	dbT2, prefixT2 := GetDb()
	if prefixT2 != prefixT {
		t.Fatalf("Test prefix after pop is %q", prefixT2)
	}
	if dbT2 == dbT {
		t.Errorf("Test db before and after pop are the same: %v", dbT2)
	}
}
