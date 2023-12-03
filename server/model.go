package server

type Message struct {
	Id     string
	Offset int
}

type PublishRequest struct {
	Body string `json:"body"`
}

type PublishResponse struct {
	Offset int `json:"offset"`
}

type GetTopicsResponse struct {
	Topics []string `json:"topics"`
}

type SubscriptionMessage struct {
	Topic     string `json:"topic"`
	MessageId string `json:"messageId"`
	Value     string `json:"value"`
}
