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

	"clickonetwo.io/whisper/internal/client"
	"clickonetwo.io/whisper/internal/conversation"
	"clickonetwo.io/whisper/internal/profile"
	"clickonetwo.io/whisper/internal/storage"
)

// SaveObjects stores all the objects in the map to the current environment
func SaveObjects(what ObjectMap) {
	var saved int
	var t string
	var as []any
	for t, as = range what {
		if len(as) > 0 {
			switch t {
			case "profiles":
				saved += saveObjects(t, as, &profile.UserProfile{})
			case "clients":
				saved += saveObjects(t, as, &client.Data{})
			case "conversations":
				saved += saveObjects(t, as, &conversation.Data{})
			case "states":
				saved += saveObjects(t, as, &conversation.State{})
			default:
				_, _ = fmt.Fprintf(os.Stderr, "Skipping objects of unknown type: %s", t)
			}
		}
	}
	if saved != 1 {
		_, _ = fmt.Fprintf(os.Stderr, "Saved %d objects.\n", saved)
	}
}

func saveObjects[T storage.StructPointer](name string, oa []any, e T) int {
	var saved int
	if len(oa) >= 10 {
		_, _ = fmt.Fprintf(os.Stderr, "Starting to save %s...", name)
	}
	for _, o := range oa {
		s, err := e.Downgrade(o)
		if err != nil {
			panic(err)
		}
		if err = storage.SaveFields(context.Background(), s); err != nil {
			panic(err)
		}
		if saved++; saved%10 == 0 {
			_, _ = fmt.Fprintf(os.Stderr, "\nSaved %d %s...", saved, name)
		}
	}
	if len(oa) >= 10 {
		_, _ = fmt.Fprintf(os.Stderr, "\n")
	}
	if saved != 1 {
		_, _ = fmt.Fprintf(os.Stderr, "Saved %d %s.\n", saved, name)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Saved 1 %s.\n", name[0:len(name)-1])
	}
	return saved
}

type ObjectMap map[string][]any

// DumpObjectsToPath serializes the entire map to the given filepath
func DumpObjectsToPath(what ObjectMap, where string) {
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
	DumpObjectsToStream(stream, what)
	if where != "-" {
		fmt.Printf("Objects dumped to %q\n", where)
	}
}

// DumpObjectsToStream marshals the objects as JSON to the given stream
func DumpObjectsToStream(stream io.Writer, what ObjectMap) {
	encoder := json.NewEncoder(stream)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(what); err != nil {
		panic(err)
	}
}

// LoadObjectsFromPath loads the objects dumped to the given filepath
func LoadObjectsFromPath(where string) ObjectMap {
	var stream io.Reader
	if where == "-" {
		stream = os.Stdin
	} else {
		if !strings.HasSuffix(where, ".json") {
			where = where + ".json"
		}
		file, err := os.OpenFile(where, os.O_RDONLY, 0o644)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		stream = file
	}
	return LoadObjectsFromStream(stream)
}

// LoadObjectsFromStream creates objects from a stream containing a JSON-serialized object map
func LoadObjectsFromStream(stream io.Reader) ObjectMap {
	decoder := json.NewDecoder(stream)
	m := make(map[string][]json.RawMessage)
	if err := decoder.Decode(&m); err != nil {
		panic(err)
	}
	om := make(ObjectMap)
	om["profiles"] = unmarshalAll("profile", profile.UserProfile{}, m["profiles"])
	om["clients"] = unmarshalAll("client", client.Data{}, m["clients"])
	om["conversations"] = unmarshalAll("conversation", conversation.Data{}, m["conversations"])
	om["states"] = unmarshalAll("state", conversation.State{}, m["states"])
	return om
}

func unmarshalAll[T any](name string, o T, ms []json.RawMessage) []any {
	objs := make([]any, 0, len(ms))
	for _, js := range ms {
		if err := json.Unmarshal(js, &o); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Skipping invalid %s data: %s", name, js)
		} else {
			objs = append(objs, any(o))
		}
	}
	return objs
}
