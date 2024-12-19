/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package console

import (
	"testing"

	"github.com/whisper-project/server.go/internal/middleware"
)

func TestPostPrefsHandler(t *testing.T) {
	r := middleware.CreateCoreEngine()
	r.POST("api/console/v0/client/prefs", PostPrefsHandler)
	err := r.Run("localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
}
