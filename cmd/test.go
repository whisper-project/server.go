/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package cmd

import (
	"context"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/whisper-project/server.golang/internal/internaltest"
	"github.com/whisper-project/server.golang/internal/storage"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Manage test data",
	Long: `This utility manages the data in the test database.
Use flags to indicate what you want to do.
Data is always loaded last.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(0)
		env, _ := cmd.Flags().GetString("env")
		err := storage.PushConfig(env)
		if err != nil {
			panic(err)
		}
		log.Printf("Operating in the %s environment.", storage.GetConfig().Name)
		var cnt int
		if cnt, _ = cmd.Flags().GetCount("clear"); cnt > 0 {
			clearDb()
		}
		if cnt, _ = cmd.Flags().GetCount("clean"); cnt > 0 {
			log.Printf("Deleting all keys created by tests...")
			internaltest.RemoveCreatedTestData()
		}
		if cnt, _ = cmd.Flags().GetCount("load"); cnt > 0 {
			loadKnownTestData()
		}
		log.Printf("Done.")
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Args = cobra.NoArgs
	testCmd.Flags().StringP("env", "e", "test", "db environment to use")
	testCmd.Flags().Count("load", "load the known test values")
	testCmd.Flags().Count("clear", "remove all values")
	testCmd.Flags().Count("clean", "remove all test-created values")
	testCmd.MarkFlagsMutuallyExclusive("clear", "clean")
	testCmd.MarkFlagsOneRequired("load", "clear", "clean")
}

func clearDb() {
	ctx := context.Background()
	db, prefix := storage.GetDb()
	log.Printf("Deleting all keys with prefix %q...", prefix)
	iter := db.Scan(ctx, 0, prefix+"*", 20).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		_ = db.Del(ctx, key)
	}
}

func loadKnownTestData() {
	log.Printf("Loading test data...")
	som, err := storage.LoadObjectsFromStream(strings.NewReader(knownTestData))
	if err != nil {
		panic(err)
	}
	om := loadObjectsFromStorage(som)
	saveObjects(om)
}

//goland:noinspection SpellCheckingInspection
var knownTestData = `{
  "clients": [
    {
      "Id": "561E5E8E-EA35-405A-A256-69C74713FAFD",
      "DeviceId": "TN+*hk9h",
      "Token": "ae032fd763a1ffa59120d4d1e96aea91de77d94942d416d5ff92e0c7780de11c",
      "LastSecret": "391c22e0641bd2655295669bd57c2570bb0cc515bc7a896d2a15d7b2a22f2715",
      "Secret": "74c69b5dab66c936b54fa47b20f453ccfa53020dd168a663e2f0c8666b087320",
      "SecretDate": 1729660449471,
      "PushId": "f01f682a-848f-4da9-91fc-268a1a70c6a8",
      "AppInfo": "mac|2.6.3",
      "UserName": "Dan Brotsky",
      "ProfileId": "B11C1B3D-21E6-4766-B16B-4FDEED785139",
      "LastLaunch": 1730163013338
    }
  ],
  "conversations": [
    {
      "Id": "3C6CE484-4A73-4D06-A8B9-4FC8EF51F5BA",
      "Name": "Anyone",
      "OwnerId": "B11C1B3D-21E6-4766-B16B-4FDEED785139",
      "StateId": ""
    }
  ],
  "profiles": [
    {
      "id": "B11C1B3D-21E6-4766-B16B-4FDEED785139",
      "lastUsed": 1730163013340,
      "name": "Dan Brotsky",
      "password": "f2efcee88b62c9143ead24aba2d159ac5e1484d21b662f89d6e743be8ef788f0",
      "whisperTimestamp": "1729383642",
      "whisperProfile": {
        "id": "B11C1B3D-21E6-4766-B16B-4FDEED785139",
        "timestamp": 1729383642,
        "defaultId": "3C6CE484-4A73-4D06-A8B9-4FC8EF51F5BA",
        "lastId": "3C6CE484-4A73-4D06-A8B9-4FC8EF51F5BA",
        "table": {
          "3C6CE484-4A73-4D06-A8B9-4FC8EF51F5BA": {
            "id": "3C6CE484-4A73-4D06-A8B9-4FC8EF51F5BA",
            "name": "Anyone",
            "allowed": {
              "0AA7D570-0063-4F91-BFA4-89C46CDB4ADB": "lisa",
              "3C3E2066-ED94-4818-BFFF-455C9091E840": "Shawna",
              "7AAB4278-3C35-46D7-82D6-2AC569C65077": "Dan in Safari",
              "87D6576A-7BDA-48F1-8837-FD53E6720BF9": "Dan 3",
              "9E2245A7-F0A8-4BD5-9A32-201A1DAB70EE": "sonya sotinsky"
            }
          },
          "45006108-22FF-4CEF-8395-24164F2EB312": {
            "id": "45006108-22FF-4CEF-8395-24164F2EB312",
            "name": "Family",
            "allowed": {
              "3C3E2066-ED94-4818-BFFF-455C9091E840": "Shawna"
            }
          },
          "C055DCC2-4783-4E04-86DD-75EDE4204B0C": {
            "id": "C055DCC2-4783-4E04-86DD-75EDE4204B0C",
            "name": "Work",
            "allowed": {
              "3C3E2066-ED94-4818-BFFF-455C9091E840": "Shawna"
            }
          }
        }
      },
      "listenTimestamp": "1729724274",
      "listenProfile": {
        "id": "B11C1B3D-21E6-4766-B16B-4FDEED785139",
        "timestamp": 1729724274,
        "table": {
          "F1B3EA28-2380-4C02-92FC-D8354905B186": {
            "id": "F1B3EA28-2380-4C02-92FC-D8354905B186",
            "name": "Chats",
            "owner": "7380CDE0-0060-4BA0-937A-D17D86EC5595",
            "ownerName": "Bill Weihl",
            "lastListened": 751417074.212161
          }
        }
      },
      "settingsVersion": 4,
      "settingsETag": "0384e7aed2e02d7eae8472c088fc31b2",
      "settingsProfile": {
        "id": "B11C1B3D-21E6-4766-B16B-4FDEED785139",
        "version": 4,
        "settings": {
          "do_server_side_transcription_preference": "yes",
          "elevenlabs_api_key_preference": "sk_ba03b28cf41881e8a695fdf26d9af2dc440b746fda526104",
          "elevenlabs_dictionary_id_preference": "",
          "elevenlabs_dictionary_version_preference": "",
          "elevenlabs_latency_reduction_preference": "1",
          "elevenlabs_voice_id_preference": "TX3LPaxmHKxFdv7VOQHJ",
          "history_buttons_preference": "r-i-f",
          "interjection_alert_preference": "",
          "interjection_prefix_preference": "",
          "listen_tap_preference": "last",
          "newest_whisper_location_preference": "bottom",
          "version": "4",
          "whisper_tap_preference": "default"
        },
        "eTag": "0384e7aed2e02d7eae8472c088fc31b2"
      },
      "favoritesTimestamp": "1726077685",
      "favoritesProfile": {
        "id": "B11C1B3D-21E6-4766-B16B-4FDEED785139",
        "timestamp": 1726077685,
        "favorites": [
          {
            "name": "Lisa Mihaly",
            "text": "Lisa Mihaly",
            "speechId": "CEe52YkgIXmp41H0fpUM",
            "speechHash": "6ln2n18efnae5"
          },
          {
            "name": "Michela Weihl",
            "text": "Michela Weihl",
            "speechId": "wQce0hGBv6Qj2QkDt2pK",
            "speechHash": "-1oobleo8u6van"
          },
          {
            "name": "Bill Weihl",
            "text": "Bill Weihl",
            "speechId": "UtKPob1kT4hnRidyMrjN",
            "speechHash": "-1oobleo8u6van"
          }
        ],
        "groupList": [
          "Numbers"
        ],
        "groupTable": {
          "Numbers": [
            "Lisa Mihaly",
            "Bill Weihl",
            "Michela Weihl"
          ]
        }
      }
    }
  ],
  "states": [
    {
      "Id": "d7dfb2b5-f25a-4de7-8c4a-52af08f1e7f3",
      "ServerId": "",
      "ConversationId": "3C6CE484-4A73-4D06-A8B9-4FC8EF51F5BA",
      "ContentId": "E7021FB4-58BF-4F24-8D31-25A4CD10E2AE",
      "TzId": "America/Los_Angeles",
      "StartTime": 1729200691559,
      "Duration": 862489,
      "ContentKey": "p:tcp:1932cb2f-bc04-4d17-87a8-5e7d19a893d2",
      "Transcription": "Trying",
      "ErrCount": 0,
      "Ttl": 0
    }
  ]
}`
