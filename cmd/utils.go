/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func DumpObjects(where string, what any, name string) {
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
