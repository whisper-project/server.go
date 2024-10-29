/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"clickonetwo.io/whisper/internal/client"
	"clickonetwo.io/whisper/internal/conversation"
	"clickonetwo.io/whisper/internal/profile"
	"clickonetwo.io/whisper/internal/storage"
)

// transferCmd represents the transfer command
var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer objects between environments",
	Long: `This command transfers whisper database objects between environments.
You must use a flag to specify which objects you want to transfer.
When transferring specific objects, you can also dump their JSON to an output file.`,
	Run: func(cmd *cobra.Command, args []string) {
		from, err := cmd.Flags().GetString("from")
		if err != nil {
			panic(err)
		}
		load, err := cmd.Flags().GetString("load")
		if err != nil {
			panic(err)
		}
		to, err := cmd.Flags().GetString("to")
		if err != nil {
			panic(err)
		}
		dump, err := cmd.Flags().GetString("dump")
		if err != nil {
			panic(err)
		}
		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			panic(err)
		}
		profiles, err := cmd.Flags().GetString("profiles")
		if err != nil {
			panic(err)
		}
		clients, err := cmd.Flags().GetString("clients")
		if err != nil {
			panic(err)
		}
		conversations, err := cmd.Flags().GetString("conversations")
		if err != nil {
			panic(err)
		}
		states, err := cmd.Flags().GetString("states")
		if err != nil {
			panic(err)
		}

		var om ObjectMap
		if from != "" {
			if err := storage.PushConfig(from); err != nil {
				panic(err)
			}
			defer storage.PopConfig()
			if all {
				om = collectAll()
			} else {
				om = make(ObjectMap)
				if ids := strings.Split(profiles, ","); profiles != "" {
					var p profile.UserProfile
					om["profiles"] = collectObjectsById("profiles", ids, &p)
				}
				if ids := strings.Split(clients, ","); clients != "" {
					var c client.Data
					om["clients"] = collectObjectsById("clients", ids, &c)
				}
				if ids := strings.Split(conversations, ","); conversations != "" {
					var c conversation.Data
					om["conversations"] = collectObjectsById("conversations", ids, &c)
				}
				if ids := strings.Split(states, ","); states != "" {
					var s conversation.State
					om["states"] = collectObjectsById("states", ids, &s)
				}
			}
		} else {
			om = LoadObjects(load)
		}
		if to != "" {
			if err := storage.PushConfig(to); err != nil {
				panic(err)
			}
			defer storage.PopConfig()
			SaveObjects(om)
		} else {
			DumpObjects(om, dump)
		}
	},
}

func init() {
	rootCmd.AddCommand(transferCmd)

	transferCmd.Flags().StringP("from", "f", "", "source environment")
	transferCmd.Flags().StringP("load", "l", "", "source file ('-' for stdin)")
	transferCmd.MarkFlagsOneRequired("from", "load")
	transferCmd.MarkFlagsMutuallyExclusive("from", "load")
	transferCmd.Flags().StringP("to", "t", "", "target environment")
	transferCmd.Flags().StringP("dump", "d", "", "output file ('-' for stdout)")
	transferCmd.MarkFlagsOneRequired("to", "dump")
	transferCmd.MarkFlagsMutuallyExclusive("to", "dump")
	transferCmd.Flags().Bool("all", false, "transfer all objects")
	transferCmd.Flags().String("profiles", "", "profile ids to transfer")
	transferCmd.Flags().String("clients", "", "client ids to transfer")
	transferCmd.Flags().String("conversations", "", "client ids to transfer")
	transferCmd.Flags().String("states", "", "state ids to transfer")
	transferCmd.MarkFlagsOneRequired("load", "all", "profiles", "clients", "conversations", "states")
	transferCmd.MarkFlagsMutuallyExclusive("load", "all", "profiles")
	transferCmd.MarkFlagsMutuallyExclusive("load", "all", "clients")
	transferCmd.MarkFlagsMutuallyExclusive("load", "all", "conversations")
	transferCmd.MarkFlagsMutuallyExclusive("load", "all", "states")
}

func collectAll() ObjectMap {
	om := make(ObjectMap)
	om["profiles"] = collectObjectsByType("profiles", &profile.UserProfile{})
	om["clients"] = collectObjectsByType("clients", &client.Data{})
	om["conversations"] = collectObjectsByType("conversations", &conversation.Data{})
	om["states"] = collectObjectsByType("states", &conversation.State{})
	return om
}

func collectObjectsByType[T storage.StorableStruct](name string, o T) []any {
	collected := 0
	var as []any
	collect := func() {
		as = append(as, o.Copy())
		if collected++; collected%10 == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "\nCollected %d %s...", collected, name)
		}
	}
	_, _ = fmt.Fprintf(os.Stderr, "Starting to collect all %s...", name)
	if err := storage.MapFields(context.Background(), collect, o); err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\nCollected %d %s.\n", len(as), name)
	return as
}

func collectObjectsById[T storage.StorableStruct](name string, ids []string, o T) []any {
	singular := name[0 : len(name)-1]
	if len(ids) >= 10 {
		_, _ = fmt.Fprintf(os.Stderr, "Starting to collect %s...", name)
	}
	as := make([]any, 0, len(ids))
	for _, id := range ids {
		_ = o.SetStorageId(id)
		if err := storage.LoadFields(context.Background(), o); err != nil {
			panic(err)
		}
		as = append(as, o.Copy())
	}
	if len(ids) >= 10 {
		_, _ = fmt.Fprintf(os.Stderr, "\n")
	}
	if len(as) != 1 {
		_, _ = fmt.Fprintf(os.Stderr, "Collected %d %s.\n", len(as), name)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Collected 1 %s.\n", singular)
	}
	return as
}
