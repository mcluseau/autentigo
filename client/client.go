package client

type Client struct {
	ServerURL string

	validationCrt []byte
}

func New(serverURL string) *Client {
	return &Client{
		ServerURL: serverURL,
	}
}
