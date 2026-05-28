package nknclient

type Client struct{}

func New(rpcURL string) *Client {
	return &Client{}
}