package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Logger = func(format string, v ...interface{})

func nooplogger(format string, v ...interface{}) {}

type Client struct {
	hookURL string
	*http.Client
	log Logger
}

func New(hookURL string, logger Logger, client *http.Client) *Client {
	if hookURL == "" {
		return nil
	}
	if logger == nil {
		logger = nooplogger
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{hookURL, client, logger}
}

func (sc *Client) Post(msg interface{}) error {
	blob, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	r := bytes.NewReader(blob)
	rsp, err := sc.Client.Post(sc.hookURL, "application/json", r)
	if err != nil {
		return err
	}
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %q", rsp.Status)
	}
	return nil
}

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type Attachment struct {
	Fallback  string  `json:"fallback"`
	Color     string  `json:"color"`
	Title     string  `json:"title"`
	TitleLink string  `json:"title_link"`
	Text      string  `json:"text"`
	TimeStamp int64   `json:"ts"`
	Fields    []Field `json:"fields"`
}

type Message struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}
