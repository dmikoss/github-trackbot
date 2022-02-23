package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type Client struct {
	host     string
	basepath string
	client   *http.Client
	offset   int
}

func NewTelegramClient(host string, token string, httpclient *http.Client) Client {
	return Client{
		host:     host,
		basepath: "bot" + token,
		client:   httpclient,
		offset:   0,
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

// delay timeout sec, max limit updates in one batch
func (c *Client) RunRecvMsgLoop(ctx context.Context, chwait chan<- struct{}, limit int, timeout int) error {
L:
	for {
		updates, err := c.updates(ctx, c.offset, limit, timeout)
		if err != nil {
			log.Println(err.Error())
		}

		for _, update := range updates {
			if c.offset < update.ID+1 {
				c.offset = update.ID + 1
			}
		}

		// TODO Send updates to message processor
		if len(updates) > 0 {
			fmt.Println(updates)
		}

		select {
		case <-ctx.Done():
			log.Println("Stopping telegram RunRecvMessages loop")
			break L
		default:
		}
	}
	chwait <- struct{}{}
	return nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int, text string) error {
	q := url.Values{}
	q.Add("chat_id", strconv.Itoa(chatID))
	q.Add("text", text)

	_, err := c.doRequest(ctx, "sendMessage", q)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) updates(ctx context.Context, offset int, limit int, timeout int) ([]Update, error) {
	q := url.Values{}
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))
	q.Add("timeout", strconv.Itoa(timeout))

	data, err := c.doRequest(ctx, "getUpdates", q)
	if err != nil {
		return nil, err
	}

	var res UpdatesResponse
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res.Result, nil
}

func (c *Client) doRequest(ctx context.Context, endpoint string, query url.Values) ([]byte, error) {
	u := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(c.basepath, endpoint),
	}

	// request preparing
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest error: %w", err)
	}
	req.URL.RawQuery = query.Encode()
	// do request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Client.Do request error: %w", err)
	}
	defer resp.Body.Close()
	// results
	body, err := io.ReadAll((resp.Body))
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll error: %w", err)
	}
	return body, nil
}
