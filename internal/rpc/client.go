package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Response struct {
	ID      int64           `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data,omitempty"`
	Message string          `json:"message"`
}

type Client struct {
	ctx    context.Context
	url    string
	secret string

	reconnectTimeout time.Duration
	reconnectEnabled bool

	conn   *websocket.Conn
	connMu sync.Mutex

	writeID atomic.Int64
	writeMu sync.Mutex

	pending   map[int64]chan Response
	pendingMu sync.Mutex
}

type ClientOption func(*Client)

func WithReconnect(enabled bool) ClientOption {
	return func(c *Client) {
		c.reconnectEnabled = enabled
	}
}

func WithReconnectTimeout(seconds int) ClientOption {
	return func(c *Client) {
		if seconds > 0 {
			c.reconnectTimeout = time.Duration(seconds) * time.Second
		}
	}
}

func NewClient(options ...ClientOption) *Client {
	client := &Client{
		reconnectTimeout: 15 * time.Second,
		reconnectEnabled: false,
		pending:          make(map[int64]chan Response),
	}

	for _, option := range options {
		option(client)
	}

	return client
}

type ConnectOption func(*Client)

func WithSecret(secret string) ConnectOption {
	return func(c *Client) {
		c.secret = secret
	}
}

func (c *Client) Connect(ctx context.Context, url string, options ...ConnectOption) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()
	c.ctx = ctx

	if c.conn != nil {
		return fmt.Errorf("RPC: Already connected")
	}

	c.url = url
	for _, option := range options {
		option(c)
	}

	if err := c.open(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Client) Reconnect(ctx context.Context) {
	slog.Info("RPC: Trying to reconnect")
	ctx = c.ctx

	for {
		if !c.wait(c.ctx) {
			return
		}

		c.connMu.Lock()
		err := c.open(c.ctx)
		c.connMu.Unlock()

		if err != nil {
			slog.Warn("RPC: Reconnect failed", "error", err)
			continue
		}

		slog.Info("RPC: Reconnected")
		return
	}
}

func (c *Client) Close(ctx context.Context) error {
	c.connMu.Lock()
	defer c.connMu.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	return err
}

type RawWrite struct {
	ID      int64  `json:"id"`
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type WriteOption func(*RawWrite)

func WithID(id int64) WriteOption {
	return func(req *RawWrite) {
		req.ID = id
	}
}

func WithParams(params any) WriteOption {
	return func(req *RawWrite) {
		req.Params = params
	}
}

func (c *Client) Write(method string, opts ...WriteOption) error {
	request := RawWrite{
		JSONRPC: "2.0",
		Method:  method,
	}

	for _, opt := range opts {
		opt(&request)
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("RPC: Not connected")
	}

	return c.conn.WriteJSON(request)
}

func (c *Client) Call(ctx context.Context, method string, params ...any) (response Response, err error) {
	id := c.writeID.Add(1)
	responseCh := make(chan Response, 1)

	c.pendingMu.Lock()
	c.pending[id] = responseCh
	c.pendingMu.Unlock()

	defer func() {
		c.pendingMu.Lock()
		delete(c.pending, id)
		c.pendingMu.Unlock()
	}()

	opts := []WriteOption{WithID(id)}
	if len(params) > 0 {
		opts = append(opts, WithParams(params[0]))
	}

	err = c.Write(method, opts...)
	if err != nil {
		return Response{}, err
	}

	select {
	case response = <-responseCh:
		if response.Error != nil {
			err = fmt.Errorf("RPC: %s", response.Error.Message)
		}
	case <-ctx.Done():
		err = fmt.Errorf("RPC: Call canceled")
	case <-c.ctx.Done():
		err = fmt.Errorf("RPC: Call canceled")
	}
	return response, err
}

func (c *Client) open(ctx context.Context) error {
	header := http.Header{}
	if c.secret != "" {
		header.Set("Authorization", "Bearer "+c.secret)
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.url, header)
	if err != nil {
		return err
	}

	c.conn = conn
	go c.run(ctx)
	return nil
}

func (c *Client) run(ctx context.Context) {
	for {
		conn := c.conn
		_, message, err := conn.ReadMessage()
		if err != nil {
			slog.Warn("RPC: Read", "error", err)
			break
		}
		var response Response
		err = json.Unmarshal([]byte(message), &response)
		slog.Debug("RPC: Message received", "message", string(message))
		if err != nil {
			return
		}
		c.pendingMu.Lock()
		ch, exists := c.pending[response.ID]
		c.pendingMu.Unlock()
		if exists {
			ch <- response
		}
		slog.Info("RPC: Message received", "ID", response.ID, "JSONRPC", response.JSONRPC, "result", response.Result.String())
	}

	c.Close(ctx)
	if c.reconnectEnabled {
		c.Reconnect(ctx)
	}
}

func (c *Client) wait(ctx context.Context) bool {
	timer := time.NewTimer(c.reconnectTimeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
