package server

type Message struct {
	Id     string
	Offset int
}

type Delivery struct {
	Topic     string `json:"topic"`
	MessageId string `json:"messageId"`
	Value     string `json:"value"`
}

type DeliveryResponse struct {
	MessageId string `json:"messageId"`
	Ack       bool   `json:"ack"`
	Err       string `json:"err"`
}

type PublishRequest struct {
	Body string `json:"body"`
}

type PublishResponse struct {
	Offset    int    `json:"offset"`
	MessageId string `json:"messageId"`
}

type GetTopicsResponse struct {
	Topics []string `json:"topics"`
}
