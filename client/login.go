package client

import "encoding/json"

type LoginResult struct {
	Token  string           `json:"token"`
	Claims *json.RawMessage `json:"claims"`
}
