package main

import "encoding/json"

type PublishRequest struct {
	Body json.RawMessage `json:"body"`
}

type PublishResponse struct {
	Offset int `json:"offset"`
}

type GetTopicsResponse struct {
	Topics []string `json:"topics"`
}
