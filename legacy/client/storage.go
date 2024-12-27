/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * GNU Affero General Public License v3, reproduced in the LICENSE file.
 */

package client

import (
	"fmt"

	"github.com/whisper-project/server.golang/common/platform"
)

type Data struct {
	Id         string `redis:"id"`
	DeviceId   string `redis:"deviceId"`
	Token      string `redis:"token"`
	LastSecret string `redis:"lastSecret"`
	Secret     string `redis:"secret"`
	SecretDate int64  `redis:"secretDate"`
	PushId     string `redis:"pushId"`
	AppInfo    string `redis:"appInfo"`
	UserName   string `redis:"userName"`
	ProfileId  string `redis:"profileId"`
	LastLaunch int64  `redis:"lastLaunch"`
}

func (d *Data) StoragePrefix() string {
	return "cli:"
}

func (d *Data) StorageId() string {
	if d == nil {
		return ""
	}
	return d.Id
}

func (d *Data) SetStorageId(id string) error {
	if d == nil {
		return fmt.Errorf("can't set platform id of nil struct")
	}
	d.Id = id
	return nil
}

func (d *Data) Copy() platform.StructPointer {
	if d == nil {
		return nil
	}
	n := new(Data)
	*n = *d
	return n
}

func (d *Data) Downgrade(in any) (platform.StructPointer, error) {
	if o, ok := in.(Data); ok {
		return &o, nil
	}
	if o, ok := in.(*Data); ok {
		return o, nil
	}
	return nil, fmt.Errorf("not a client.Data: %#v", in)
}
