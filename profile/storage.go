package profile

type Data struct {
	Id                 string `redis:"id"`
	Name               string `redis:"name"`
	Password           string `redis:"password"`
	WhisperTimestamp   string `redis:"whisperTimestamp"`
	WhisperProfile     string `redis:"whisperProfile"`
	ListenTimestamp    string `redis:"listenTimestamp"`
	ListenProfile      string `redis:"listenProfile"`
	SettingsVersion    int64  `redis:"settingsVersion"`
	SettingsETag       string `redis:"settingsETag"`
	SettingsProfile    string `redis:"settingsProfile"`
	FavoritesTimestamp string `redis:"favoritesTimestamp"`
	FavoritesProfile   string `redis:"favoritesProfile"`
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
