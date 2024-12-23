/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package apns

type Storage struct {
	id        string
	clientKey string
	status    int64
	devId     string
	reason    string
	timestamp int64
}
