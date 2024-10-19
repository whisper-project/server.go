/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"clickonetwo.io/whisper/internal/client"
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
		to, err := cmd.Flags().GetString("to")
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
		dump, err := cmd.Flags().GetString("dump")
		if all {
			transferAll(from, to)
		} else if profiles != "" {
			ps := transferProfiles(from, to, profiles)
			if dump != "" {
				dumpObjects(dump, ps, "Profiles")
			}
		} else if clients != "" {
			cs := transferClients(from, to, clients)
			if dump != "" {
				dumpObjects(dump, cs, "Clients")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(transferCmd)

	transferCmd.Flags().StringP("from", "f", "production", "source environment")
	transferCmd.Flags().StringP("to", "t", "development", "target environment")
	transferCmd.Flags().BoolP("all", "a", false, "transfer all objects")
	transferCmd.Flags().StringP("profiles", "p", "", "profile ids to transfer")
	transferCmd.Flags().StringP("clients", "c", "", "client ids to transfer")
	transferCmd.MarkFlagsOneRequired("all", "profiles", "clients")
	transferCmd.MarkFlagsMutuallyExclusive("all", "profiles", "clients")
	transferCmd.Flags().StringP("dump", "d", "", "JSON output file ('-' for stdout)")
}

func transferAll(from, to string) {
	_, _ = fmt.Fprintf(os.Stderr, "Transferring all data from %s to %s...\n", from, to)
	if err := storage.PushConfig(from); err != nil {
		panic(err)
	}
	ps, cs, cls := collectAll()
	storage.PopConfig()
	if err := storage.PushConfig(to); err != nil {
		panic(err)
	}
	saveAll(ps, cs, cls)
	storage.PopConfig()
}

func transferProfiles(from, to string, profiles string) []profile.UserProfile {
	_, _ = fmt.Fprintf(os.Stderr, "Transferring profiles from %s to %s...\n", from, to)
	if err := storage.PushConfig(from); err != nil {
		panic(err)
	}
	ids := strings.Split(profiles, ",")
	ps := collectProfiles(ids)
	storage.PopConfig()
	if err := storage.PushConfig(to); err != nil {
		panic(err)
	}
	saveProfiles(ps, nil)
	storage.PopConfig()
	return ps
}

func transferClients(from, to string, clients string) []client.Data {
	_, _ = fmt.Fprintf(os.Stderr, "Transferring clients from %s to %s...\n", from, to)
	if err := storage.PushConfig(from); err != nil {
		panic(err)
	}
	ids := strings.Split(clients, ",")
	cs := collectClients(ids)
	storage.PopConfig()
	if err := storage.PushConfig(to); err != nil {
		panic(err)
	}
	saveClients(cs)
	storage.PopConfig()
	return cs
}

func collectAll() (ps []profile.UserProfile, cs []client.Data, cls map[string][]string) {
	ps, cls = collectAllProfiles()
	cs = collectAllClients()
	return
}

func collectAllProfiles() (ps []profile.UserProfile, cls map[string][]string) {
	cls = make(map[string][]string)
	p := &profile.UserProfile{}
	count := 0
	pushProfile := func() {
		ps = append(ps, *p)
		cl, err := storage.FetchMembers(context.Background(), profile.Clients(p.Id))
		if err != nil {
			panic(err)
		}
		cls[p.Id] = cl
		count++
		if (count % 10) == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "\nCollected %d profiles...", count)
		}
	}

	_, _ = fmt.Fprintf(os.Stderr, "Starting to collect profiles...")
	if err := storage.MapFields(context.Background(), pushProfile, p); err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\nCollected %d profiles.\n", count)
	return
}

func collectProfiles(ids []string) []profile.UserProfile {
	_, _ = fmt.Fprintf(os.Stderr, "Starting to collect profiles...")
	ps := make([]profile.UserProfile, 0, len(ids))
	for _, pid := range ids {
		p := profile.UserProfile{Id: pid}
		if err := storage.LoadFields(context.Background(), &p); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "\n...error loading profile %s: %s", pid, err)
		}
		ps = append(ps, p)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\nCollected %d profile(s).\n", len(ps))
	return ps
}

func collectAllClients() (cs []client.Data) {
	c := &client.Data{}
	count := 0
	pushProfile := func() {
		cs = append(cs, *c)
		count++
		if (count % 10) == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "\nCollected %d clients...", count)
		}
	}

	_, _ = fmt.Fprintf(os.Stderr, "Starting to collect clients...")
	if err := storage.MapFields(context.Background(), pushProfile, c); err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\nCollected %d clients.\n", count)
	return
}

func collectClients(ids []string) []client.Data {
	_, _ = fmt.Fprintf(os.Stderr, "Starting to collect clients...")
	cs := make([]client.Data, 0, len(ids))
	for _, cid := range ids {
		c := client.Data{Id: cid}
		if err := storage.LoadFields(context.Background(), &c); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "\n...error loading client %s: %s", cid, err)
		}
		cs = append(cs, c)
	}
	_, _ = fmt.Fprintf(os.Stderr, "\nCollected %d client(s).\n", len(cs))
	return cs
}

func saveAll(ps []profile.UserProfile, cs []client.Data, cls map[string][]string) {
	saveProfiles(ps, cls)
	saveClients(cs)
}

func saveProfiles(ps []profile.UserProfile, cls map[string][]string) {
	_, _ = fmt.Fprintf(os.Stderr, "Starting to save profiles...")
	count := 0
	for _, p := range ps {
		if err := storage.SaveFields(context.Background(), &p); err != nil {
			panic(err)
		}
		if err := storage.AddMembers(context.Background(), profile.Clients(p.Id), cls[p.Id]...); err != nil {
			panic(err)
		}
		count++
		if (count % 10) == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "\nSaved %d profiles...", count)
		}
	}
	_, _ = fmt.Fprintf(os.Stderr, "\nSaved %d profile(s).\n", len(ps))
}

func saveClients(cs []client.Data) {
	_, _ = fmt.Fprintf(os.Stderr, "Starting to save clients...")
	count := 0
	for _, tp := range cs {
		if err := storage.SaveFields(context.Background(), &tp); err != nil {
			panic(err)
		}
		count++
		if (count % 10) == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "\nSaved %d clients...", count)
		}
	}
	_, _ = fmt.Fprintf(os.Stderr, "\nSaved %d client(s).\n", len(cs))
}

func dumpObjects(where string, what any, name string) {
	var stream io.Writer
	if where == "-" {
		stream = os.Stdout
	} else {
		if !strings.HasSuffix(where, ".json") {
			where = where + ".json"
		}
		file, err := os.OpenFile(where, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		stream = file
	}
	encoder := json.NewEncoder(stream)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(what); err != nil {
		panic(err)
	}
	if where != "-" {
		fmt.Printf("%s dumped to %q\n", name, where)
	}
}
