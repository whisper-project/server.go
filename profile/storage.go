package profile

type Data struct {
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

func (p Data) StoragePrefix() string {
	return "pro:"
}

func (p Data) StorageId() string {
	return p.Id
}

type ClientList string

func (pcl ClientList) StoragePrefix() string {
	return "pro-clients:"
}

func (pcl ClientList) StorageId() string {
	return string(pcl)
}
