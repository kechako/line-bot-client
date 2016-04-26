package line

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const (
	EventTypeMessage     = "138311609000106303"
	EventTypeOperation   = "138311609100106403"
	EventTypeSendMessage = "138311608800106203"
)

const (
	ContentTypeText = 1 + iota
	ContentTypeImage
	ContentTypeVideo
	ContentTypeAudio
	_
	_
	ContentTypeLocation
	ContentTypeSticker
	ContentTypeContact
)

const (
	ToChannel = 1383378250
)

const (
	ContentGetUrl = "https://trialbot-api.line.me/v1/bot/message/%s/content"
)

type Request struct {
	Result []*Result `json:result`
}

type Result struct {
	From        string          `json:"from"`
	FromChannel int             `json:"fromChannel"`
	To          []string        `json:"to"`
	ToChannel   int             `json:"toChannel"`
	EventType   string          `json:"eventType"`
	Id          string          `json:"id"`
	CreatedTime int             `json:"createdTime"`
	Content     json.RawMessage `json:"content"`
	Message     *Message        `json:"-"`
	Operation   *Operation      `json:"-"`
}

type Message struct {
	Id              string          `json:"id"`
	ContentType     int             `json:"contentType"`
	From            string          `json:"from"`
	CreatedTime     int             `json:"createdTime"`
	To              []string        `json:"to"`
	ToType          int             `json:"toType"`
	ContentMetadata json.RawMessage `json:"contentMetadata"`
	Text            string          `json:"text"`
	Location        Location        `json:"location"`
}

type Location struct {
	Title     string  `json:"title"`
	Address   string  `json:"address"`
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

type Operation struct {
	Revision int      `json:"revision"`
	OpType   int      `json:"opType"`
	params   []string `json:"params"`
}

func ParseRequest(r io.Reader) (*Request, error) {
	req := &Request{}
	err := json.NewDecoder(r).Decode(req)
	if err != nil {
		return nil, errors.Wrap(err, "Can not decode request json")
	}

	for _, res := range req.Result {
		switch res.EventType {
		case EventTypeMessage:
			msg := &Message{}
			err = json.Unmarshal(res.Content, msg)
			if err != nil {
				return nil, errors.Wrap(err, "Can not decode message json")
			}
			res.Message = msg
		case EventTypeOperation:
			op := &Operation{}
			err = json.Unmarshal(res.Content, op)
			if err != nil {
				return nil, errors.Wrap(err, "Can not decode operation json")
			}
			res.Operation = op
		}
	}

	return req, nil
}

func (m *Message) GetContent(client *http.Client) (*http.Response, error) {
	url := fmt.Sprintf(ContentGetUrl, m.Id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Can not create a new request")
	}

	req.Header.Set("X-Line-ChannelID", LineChannelID)
	req.Header.Set("X-Line-ChannelSecret", LineChannelSecret)
	req.Header.Set("X-Line-Trusted-User-With-ACL", MID)

	return client.Do(req)
}
