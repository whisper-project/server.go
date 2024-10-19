/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package storage

import "testing"

func TestGetConfig(t *testing.T) {
	if GetConfig() != testConfig {
		t.Errorf("Initial configuration is not the test configuration")
	}
}

func TestPushPopConfig(t *testing.T) {
	popTest := func() {
		PopConfig()
		if GetConfig() != testConfig {
			t.Errorf("Config after pop is not the test configuration")
		}
	}
	if err := PushConfig("staging"); err != nil {
		t.Errorf("Failed to push config and load staging configuration")
	}
	defer popTest()
	if GetConfig() == testConfig {
		t.Errorf("Config after push is still the test configuration")
	}
}

func TestPushPopFailedConfig(t *testing.T) {
	if err := PushConfig(".no-such-environment-file"); err == nil {
		t.Errorf("Was able to push a non-existent environment")
	}
	defer PopConfig()
	if GetConfig() != testConfig {
		t.Errorf("Config after failed push is not the test configuration")
	}
}

func TestMultiPushPopConfig(t *testing.T) {
	configT := GetConfig()
	if configT.DbKeyPrefix != "t:" {
		t.Errorf("Initial config prefix is wrong: %q", configT.DbKeyPrefix)
	}
	if err := PushConfig("development"); err != nil {
		t.Fatalf("failed to push development config: %v", err)
	}
	configD := GetConfig()
	if configT == configD {
		t.Errorf("Configs before and after dev push are the same")
	}
	if configD.DbKeyPrefix != "d:" {
		t.Errorf("Prefix after dev push is wrong: %q", configD.DbKeyPrefix)
	}
	if err := PushConfig("staging"); err != nil {
		t.Fatalf("failed to push staging config: %v", err)
	}
	configS := GetConfig()
	if configD == configS {
		t.Errorf("Dbs before and after stage push are the same")
	}
	if configS.DbKeyPrefix != "s:" {
		t.Errorf("Prefix after stage push is wrong: %q", configS.DbKeyPrefix)
	}
	PopConfig()
	configD2 := GetConfig()
	if configD2 != configD {
		t.Errorf("Dev config before and after pop are different")
	}
	if err := PushConfig("production"); err != nil {
		t.Fatalf("failed to push production config: %v", err)
	}
	configP := GetConfig()
	if configP == configD2 {
		t.Errorf("Configs before and after prod push are the same")
	}
	if configP.DbKeyPrefix != "p:" {
		t.Errorf("Prefix after prod push is wrong: %q", configP.DbKeyPrefix)
	}
	PopConfig()
	configD3 := GetConfig()
	if configD3 != configD2 {
		t.Errorf("Dev config before and after second pop are different")
	}
	PopConfig()
	configT2 := GetConfig()
	if configT2 != configT {
		t.Errorf("Test config before and after pop are different")
	}
}
