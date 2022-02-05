package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type Client struct {
	host     string
	basepath string
	client   http.Client
}

func NewTelegramClient(host string, token string) Client {
	return Client{
		host:     host,
		basepath: "bot" + token,
		client:   http.Client{},
	}
}

/*
	structs for Telegram api
*/

type Message struct {
	ID   int    `json:"message_id"`
	Text string `json:"text"`
}

type Update struct {
	ID      int     `json:"update_id"`
	Message Message `json:"message"`
}

type UpdatesResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

func (c *Client) Updates(offset int, limit int) ([]Update, error) {
	q := url.Values{}
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	data, err := c.DoRequest("getUpdates", q)
	if err != nil {
		return nil, err
	}

	var res UpdatesResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res.Result, nil
}

func (c *Client) SendMessage(chatID int, text string) error {
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatID))
	q.Add("text", text)

	_, err := c.DoRequest("sendMessage", q)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DoRequest(method string, query url.Values) ([]byte, error) {
	u := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(c.basepath, method),
	}

	// request preparing
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cant do request: %w", err)
	}
	req.URL.RawQuery = query.Encode()
	// do request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cant do request: %w", err)
	}
	defer resp.Body.Close()
	// results
	body, err := io.ReadAll((resp.Body))
	if err != nil {
		return nil, fmt.Errorf("cant do request: %w", err)
	}
	return body, nil
}
