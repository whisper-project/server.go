package storage

type ProfileData struct {
	Id                 string
	Name               string
	Password           string
	WhisperTimestamp   string
	WhisperProfile     string
	ListenTimestamp    string
	ListenProfile      string
	SettingsVersion    int64
	SettingsETag       string
	SettingsProfile    string
	FavoritesTimestamp string
	FavoritesProfile   string
}

func (p ProfileData) prefix() string {
	return "pro:"
}

func (p ProfileData) id() string {
	return p.Id
}

type ProfileClientList string

func (pcl ProfileClientList) prefix() string {
	return "pro-clients:"
}

func (pcl ProfileClientList) id() string {
	return string(pcl)
}
