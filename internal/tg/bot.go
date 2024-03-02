// Package tg implements a barebone Telegram API client.
package tg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const apiBase = "https://api.telegram.org"

// Bot is a client to interact with the Telegram Bot API.
// Must be constructed using [NewBot].
type Bot struct {
	token      string
	httpClient *http.Client
}

// NewBot constructs a new Bot object.
func NewBot(token string) *Bot {
	return &Bot{
		token: token,
		httpClient: &http.Client{
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// Represents a response produced by Telegram API.
type apiResponse[ResultType any] struct {
	Ok          bool       `json:"ok"`
	Description string     `json:"description,omitempty"`
	Result      ResultType `json:"result,omitempty"`
	ErrorCode   int        `json:"error_code,omitempty"`
}

func (bot *Bot) apiRequest(ctx context.Context, apiMethod string, params url.Values) ([]byte, error) {
	apiURL := apiBase + "/bot" + bot.token + "/" + apiMethod
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := bot.httpClient.Do(req)
	if err != nil {
		// Prevents the URL from popping up in the error, which would
		// otherwise leak the bot's token.
		var ue *url.Error
		if errors.As(err, &ue) {
			ue.URL = "<URL hidden>"
		}
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Represents an error produced by Telegram API.
type APIError struct {
	ErrorCode   int
	Description string
}

func (e APIError) Error() string {
	return fmt.Sprintf("api error %d: \"%s\"", e.ErrorCode, e.Description)
}

func intoResult[ResultType any](rawResp []byte) (ResultType, error) {
	var resp apiResponse[ResultType]
	if err := json.Unmarshal(rawResp, &resp); err != nil {
		return *new(ResultType), err
	}

	if !resp.Ok {
		return *new(ResultType), APIError{ErrorCode: resp.ErrorCode, Description: resp.Description}
	}

	return resp.Result, nil
}

// https://core.telegram.org/bots/api#getme
func (bot *Bot) GetMe(ctx context.Context) (*User, error) {
	rawResp, err := bot.apiRequest(ctx, "getMe", url.Values{})
	if err != nil {
		return nil, err
	}

	return intoResult[*User](rawResp)
}

// https://core.telegram.org/bots/api#getupdates
func (bot *Bot) GetUpdates(ctx context.Context, offset int64, timeout int) ([]*Update, error) {
	params := url.Values{}
	params.Set("offset", strconv.FormatInt(offset, 10))
	params.Set("timeout", strconv.Itoa(timeout))

	rawResp, err := bot.apiRequest(ctx, "getUpdates", params)
	if err != nil {
		return nil, err
	}

	return intoResult[[]*Update](rawResp)
}

type SendMessageOpts struct {
	ParseMode string
}

func (o *SendMessageOpts) into(v url.Values) {
	v.Set("parse_mode", o.ParseMode)
}

// https://core.telegram.org/bots/api#sendmessage
func (bot *Bot) SendMessage(ctx context.Context, chatID int64, text string, opts *SendMessageOpts) (*Message, error) {
	params := url.Values{}
	params.Set("chat_id", strconv.FormatInt(chatID, 10))
	params.Set("text", text)
	if opts != nil {
		opts.into(params)
	}

	rawResp, err := bot.apiRequest(ctx, "sendMessage", params)
	if err != nil {
		return nil, err
	}

	return intoResult[*Message](rawResp)
}

// https://core.telegram.org/bots/api#approvechatjoinrequest
func (bot *Bot) ApproveChatJoinRequest(ctx context.Context, chatID int64, userID int64) (bool, error) {
	params := url.Values{}
	params.Set("chat_id", strconv.FormatInt(chatID, 10))
	params.Set("user_id", strconv.FormatInt(userID, 10))

	rawResp, err := bot.apiRequest(ctx, "approveChatJoinRequest", params)
	if err != nil {
		return false, err
	}

	return intoResult[bool](rawResp)
}

// https://core.telegram.org/bots/api#declinechatjoinrequest
func (bot *Bot) DeclineChatJoinRequest(ctx context.Context, chatID int64, userID int64) (bool, error) {
	params := url.Values{}
	params.Set("chat_id", strconv.FormatInt(chatID, 10))
	params.Set("user_id", strconv.FormatInt(userID, 10))

	rawResp, err := bot.apiRequest(ctx, "declineChatJoinRequest", params)
	if err != nil {
		return false, err
	}

	return intoResult[bool](rawResp)
}

// https://core.telegram.org/bots/api#getchatmember
func (bot *Bot) GetChatMember(ctx context.Context, chatID int64, userID int64) (*ChatMember, error) {
	params := url.Values{}
	params.Set("chat_id", strconv.FormatInt(chatID, 10))
	params.Set("user_id", strconv.FormatInt(userID, 10))

	rawResp, err := bot.apiRequest(ctx, "getChatMember", params)
	if err != nil {
		return nil, err
	}

	return intoResult[*ChatMember](rawResp)
}
