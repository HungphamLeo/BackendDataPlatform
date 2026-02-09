package xtb

import "time"

// Config defines how we connect to XTB xAPI.
//
// In XTB, there are 2 separate connections:
//   1) API (commandExecute / execute): login, getChartRangeRequest, ...
//   2) Streaming: subscribe via streamSessionId (getTickPrices, getBalance, ...)
//
// This mirrors xtb_connection.txt (Python) so the Go code keeps the same contract.
type Config struct {
	Address       string
	APIPort       int
	StreamingPort int
	TLSEnabled    bool

	// MaxConnTries controls reconnect attempts on connect.
	MaxConnTries int
	// DialTimeout controls dial timeout.
	DialTimeout time.Duration
	// SendDelay is an optional throttling delay between socket sends.
	SendDelay time.Duration
	// ReadBufferSize used when reading from socket.
	ReadBufferSize int
}

func DefaultConfig() Config {
	return Config{
		Address:        "xapi.xtb.com",
		APIPort:        5124, // demo by default
		StreamingPort:  5125,
		TLSEnabled:     true,
		MaxConnTries:   5,
		DialTimeout:    10 * time.Second,
		SendDelay:      100 * time.Millisecond,
		ReadBufferSize: 4096,
	}
}
