/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package profile

type UploadedWhisperProfile WhisperProfile

type UploadedListenProfile ListenProfile

type UploadedAppSettings struct {
	Id       string `json:"id"`
	Version  int64  `json:"version"`
	Settings string `json:"settings"`
	ETag     string `json:"etag"`
}

type UploadedAppFavorites AppFavorites
