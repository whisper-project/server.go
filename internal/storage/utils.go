/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package storage

func SetIfMissing[T int64 | float64 | string | bool](loc *T, val T) {
	if *loc == *new(T) {
		*loc = val
	}
}
