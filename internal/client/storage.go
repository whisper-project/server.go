/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package client

type Data struct {
	Id                string `redis:"id"`
	DeviceId          string `redis:"deviceId"`
	Token             string `redis:"token"`
	LastSecret        string `redis:"lastSecret"`
	Secret            string `redis:"secret"`
	SecretDate        int64  `redis:"secretDate"`
	PushId            string `redis:"pushId"`
	AppInfo           string `redis:"appInfo"`
	UserName          string `redis:"userName"`
	ProfileId         string `redis:"profileId"`
	LastLaunch        int64  `redis:"lastLaunch"`
	IsPresenceLogging int64  `redis:"isPresenceLogging"`
}

func (data Data) StoragePrefix() string {
	return "cli:"
}

func (data Data) StorageId() string {
	return data.Id
}
