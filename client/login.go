package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type LoginResult struct {
	Token  string           `json:"token"`
	Claims *json.RawMessage `json:"claims"`
}

func (c *Client) Login(username, password string) (result LoginResult, err error) {
	loginReq, err := http.NewRequest("GET", c.ServerURL+"/basic", nil)
	if err != nil {
		return
	}
	loginReq.SetBasicAuth(username, password)

	res, err := http.DefaultClient.Do(loginReq)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		err = fmt.Errorf("login request failed with status code %d: %s", res.StatusCode, res.Status)
		return
	}

	err = json.NewDecoder(res.Body).Decode(&result)
	return
}
