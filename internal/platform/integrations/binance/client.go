package websocket

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
}

func NewClient(url string) (*Client, error) {

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		return nil, err
	}

	return &Client{conn}, nil
}

func (c *Client) Read(handler func([]byte)) {

	for {

		_, message, err := c.conn.ReadMessage()

		if err != nil {
			log.Println("Read error:", err)
			return
		}

		handler(message)
	}
}

func (c *Client) Close() {
	c.conn.Close()
}
