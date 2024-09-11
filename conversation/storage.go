package conversation

type Data struct {
	Id      string `redis:"id"`
	Name    string `redis:"name"`
	OwnerId string `redis:"ownerId"`
	StateId string `redis:"stateId"`
}

func (c Data) StoragePrefix() string {
	return "con:"
}

func (c Data) StorageId() string {
	return c.Id
}

type State struct {
	Id             string `redis:"id"`
	ServerId       string `redis:"serverId"`
	ConversationId string `redis:"conversationId"`
	ContentId      string `redis:"contentId"`
	TzId           string `redis:"tzId"`
	StartTime      int64  `redis:"startTime"`
	Duration       int64  `redis:"duration"`
	ContentKey     string `redis:"contentKey"`
	Transcription  string `redis:"transcription"`
	ErrCount       int64  `redis:"errCount"`
	Ttl            int64  `redis:"ttl"`
}

func (c State) StoragePrefix() string {
	return "tra:"
}

func (c State) StorageId() string {
	return c.Id
}
