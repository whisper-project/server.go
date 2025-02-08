/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package platform

import (
	"crypto/sha1"
	"encoding/base64"
)

func SetIfMissing[T int64 | float64 | string | bool](loc *T, val T) {
	if *loc == *new(T) {
		*loc = val
	}
}

// from https://stackoverflow.com/a/10701951/558006
func MakeSha1(s string) string {
	hashFn := sha1.New()
	hashFn.Write([]byte(s))
	return base64.URLEncoding.EncodeToString(hashFn.Sum(nil))
}
