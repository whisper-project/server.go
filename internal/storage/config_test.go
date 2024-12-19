/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package storage

import (
	"os"
	"testing"

	"github.com/go-test/deep"
)

func TestGetConfig(t *testing.T) {
	if GetConfig() != ciConfig {
		t.Errorf("Initial configuration is not the CI configuration")
	}
}

func TestPushAlteredConfig(t *testing.T) {
	env := GetConfig()
	env.Name = "AlteredTestConfig"
	PushAlteredConfig(env)
	env1 := GetConfig()
	if diff := deep.Equal(env, env1); diff != nil {
		t.Error(diff)
	}
	PopConfig()
	env2 := GetConfig()
	if diff := deep.Equal(env, env2); diff == nil {
		t.Errorf("Popped configuration is the altered conifiguration")
	}
	if env2 != ciConfig {
		t.Errorf("Popped configuration is not the original configuration")
	}
}

func TestPushPopConfig(t *testing.T) {
	if GetConfig() != ciConfig {
		t.Errorf("Initial configuration is not the CI configuration")
	}
	popTest := func() {
		PopConfig()
		if GetConfig() != ciConfig {
			t.Errorf("Environment after pop is not the CI configuration")
		}
	}
	if err := PushConfig("staging"); err != nil {
		t.Errorf("Failed to push config and load staging configuration")
	}
	defer popTest()
	if GetConfig() == ciConfig {
		t.Errorf("Environment after push is still the CI configuration")
	}
}

func TestPushPopFailedConfig(t *testing.T) {
	if GetConfig() != ciConfig {
		t.Errorf("Initial configuration is not the CI configuration")
	}
	if err := PushConfig(".no-such-environment-file"); err == nil {
		t.Errorf("Was able to push a non-existent environment")
	}
	defer PopConfig()
	if GetConfig() != ciConfig {
		t.Errorf("Environment after failed push is not the CI configuration")
	}
}

func TestPushTestConfig(t *testing.T) {
	if err := PushConfig("testing"); err != nil {
		t.Fatalf("failed to push testing config: %v", err)
	}
	defer PopConfig()
	if cfg := GetConfig().DbKeyPrefix; cfg != "t:" {
		t.Errorf("Prefix after test push is wrong: %q", cfg)
	}
}

func TestMultiPushPopConfig(t *testing.T) {
	if GetConfig() != ciConfig {
		t.Errorf("Initial configuration is not the CI configuration")
	}
	configC := GetConfig()
	if configC.DbKeyPrefix != "c:" {
		t.Errorf("Initial config prefix is wrong: %q", configC.DbKeyPrefix)
	}
	if err := PushConfig("development"); err != nil {
		t.Fatalf("failed to push development config: %v", err)
	}
	configD := GetConfig()
	if configC == configD {
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
	if configT2 != configC {
		t.Errorf("Test config before and after pop are different")
	}
	if GetConfig() != ciConfig {
		t.Errorf("Terminal configuration is not the CI configuration")
	}
}

func TestPushPopPopTestConfig(t *testing.T) {
	if err := PushConfig("ci"); err != nil {
		t.Fatalf("Failed to push configuration: %v", err)
	}
	if GetConfig() != ciConfig {
		t.Errorf("Pushed configuration is not the CI configuration")
	}
	PopConfig()
	if GetConfig() != ciConfig {
		t.Errorf("Popped configuration is not the CI configuration")
	}
	PopConfig()
	if GetConfig() != ciConfig {
		t.Errorf("Overpopped configuration is not the CI configuration")
	}
}

func TestPushVaultConfig(t *testing.T) {
	if GetConfig() != ciConfig {
		t.Errorf("Initial configuration is not the CI configuration")
	}
	var o, n string
	var err error
	// because we have no environment support, the vault load will load the unencrypted (dev) environment
	if err = PushConfig("development"); err != nil {
		t.Fatalf("Failed to push production configuration: %v", err)
	}
	defer PopConfig()
	prodConfig := GetConfig()
	if o, err = os.Getwd(); err != nil {
		t.Fatalf("Failed to get before directory: %v", err)
	}
	if err = PushConfig(""); err != nil {
		t.Fatalf("Failed to push encrypted configuration: %v", err)
	}
	defer PopConfig()
	if n, err = os.Getwd(); err != nil {
		t.Fatalf("Failed to get after directory: %v", err)
	}
	if n != o {
		t.Errorf("Directory after (%s) not same as directory before (%s) push", n, o)
	}
	encryptedConfig := GetConfig()
	if diff := deep.Equal(prodConfig, encryptedConfig); diff != nil {
		t.Errorf("Pushed encrypted configuration doesn't match the pushed production configuration: %v", diff)
	}
}

func TestFindEnvFile(t *testing.T) {
	if GetConfig() != ciConfig {
		t.Errorf("Initial configuration is not the CI configuration")
	}
	if _, err := FindEnvFile(".env.no-such-environment-file", false); err == nil {
		t.Errorf("Didn't err when file didn't exist in parent")
	}
	if _, err := FindEnvFile(".env.no-such-environment-file", true); err != nil {
		t.Errorf("Didn't find fallback when file didn't exist in parent")
	}
	if d, err := FindEnvFile(".env.vault", false); err != nil {
		t.Errorf("Didn't find .env.vault in parent")
	} else {
		if d != "../../" {
			t.Errorf("Found .env.vault in wrong parent")
		}
	}
}

func TestFindVaultLocally(t *testing.T) {
	c, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	d, err := FindEnvFile(".env.vault", false)
	if err != nil {
		t.Fatalf("Didn't find .env.vault in parent")
	}
	if err = os.Chdir(d); err != nil {
		t.Fatalf("Failed to chdir into parent dir: %v", err)
	}
	if err = pushEnvConfig(""); err != nil {
		t.Errorf("Failed to find local .env.vault: %v", err)
	}
	defer PopConfig()
	if err = os.Chdir(c); err != nil {
		t.Fatalf("Failed to return to child directory: %v", err)
	}
}
