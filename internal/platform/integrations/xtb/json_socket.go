package xtb

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// jsonSocket implements the JsonSocket behavior from xtb_connection.txt.
// It can decode sequential JSON objects from a TCP stream (no explicit delimiter)
// using json.Decoder.
type jsonSocket struct {
	cfg  Config
	port int

	mu   sync.Mutex
	conn net.Conn
	dec  *json.Decoder
}

func newJSONSocket(cfg Config, port int) *jsonSocket {
	return &jsonSocket{cfg: cfg, port: port}
}

func (s *jsonSocket) Connect(ctx context.Context) error {
	var lastErr error
	for i := 0; i < s.cfg.MaxConnTries; i++ {
		d := net.Dialer{Timeout: s.cfg.DialTimeout}
		addr := fmt.Sprintf("%s:%d", s.cfg.Address, s.port)

		var c net.Conn
		var err error
		if s.cfg.TLSEnabled {
			c, err = tls.DialWithDialer(&d, "tcp", addr, &tls.Config{ServerName: s.cfg.Address})
		} else {
			c, err = d.DialContext(ctx, "tcp", addr)
		}
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}

		s.mu.Lock()
		s.conn = c
		br := bufio.NewReaderSize(c, s.cfg.ReadBufferSize)
		s.dec = json.NewDecoder(br)
		s.dec.UseNumber()
		s.mu.Unlock()
		return nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("unable to connect")
	}
	return fmt.Errorf("xtb jsonSocket connect failed after %d tries: %w", s.cfg.MaxConnTries, lastErr)
}

func (s *jsonSocket) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		s.dec = nil
		return err
	}
	return nil
}

func (s *jsonSocket) Send(obj any) error {
	s.mu.Lock()
	c := s.conn
	s.mu.Unlock()
	if c == nil {
		return fmt.Errorf("xtb socket is not connected")
	}

	payload, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	// XTB server accepts JSON objects concatenated, same as Python implementation.
	_, err = c.Write(payload)
	if err != nil {
		return err
	}

	if s.cfg.SendDelay > 0 {
		time.Sleep(s.cfg.SendDelay)
	}
	return nil
}

// ReadOne blocks until it can decode exactly one JSON object.
func (s *jsonSocket) ReadOne() (map[string]any, error) {
	s.mu.Lock()
	dec := s.dec
	s.mu.Unlock()
	if dec == nil {
		return nil, fmt.Errorf("xtb socket is not connected")
	}
	var m map[string]any
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}
