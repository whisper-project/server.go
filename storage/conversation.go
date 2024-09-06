package storage

type ConversationData struct {
	Id      string `redis:"id"`
	Name    string `redis:"name"`
	OwnerId string `redis:"ownerId"`
	StateId string `redis:"stateId"`
}

func (c ConversationData) prefix() string {
	return "con:"
}

func (c ConversationData) id() string {
	return c.Id
}

type ConversationState struct {
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

func (c ConversationState) prefix() string {
	return "tra:"
}

func (c ConversationState) id() string {
	return c.Id
}
