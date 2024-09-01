package storage

import (
	"context"
)

type ClientData struct {
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

func (data ClientData) prefix() string {
	return "cli:"
}

func (data ClientData) id() string {
	return data.Id
}

func (received *ClientData) HasChanged(ctx context.Context) (bool, string) {
	existing := &ClientData{Id: received.Id}
	if err := LoadFields(ctx, existing); err != nil {
		return true, "APNS token from new"
	}
	if existing.LastSecret != received.LastSecret {
		return true, "unconfirmed secret from existing"
	}
	if existing.Token != received.Token {
		return true, "new APNS token from existing"
	}
	if existing.AppInfo != received.AppInfo {
		return true, "new build data from existing"
	}
	if received.IsPresenceLogging == 0 {
		return true, "no presence logging from existing"
	}
	return false, ""
}
