package toot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/quells/mastobot/internal/app"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Visibility int

const (
	VisibilityInvalid Visibility = iota
	VisibilityPrivate
	VisibilityUnlisted
	VisibilityPublic
	VisibilityDirect
)

func (v Visibility) String() string {
	switch v {
	case VisibilityPrivate:
		return "private"
	case VisibilityUnlisted:
		return "unlisted"
	case VisibilityPublic:
		return "public"
	case VisibilityDirect:
		return "direct"
	default:
		return ""
	}
}

func VisibilityFrom(s string) Visibility {
	switch strings.ToLower(s) {
	case "private":
		return VisibilityPrivate
	case "unlisted":
		return VisibilityUnlisted
	case "public":
		return VisibilityPublic
	case "direct":
		return VisibilityDirect
	default:
		return VisibilityInvalid
	}
}

type Status struct {
	Text       string
	MediaIDs   []string
	ReplyToID  string
	Sensitive  bool
	Spoiler    string
	Visibility Visibility
}

func (s Status) FormData() url.Values {
	f := url.Values{
		"status":     []string{s.Text},
		"visibility": []string{s.Visibility.String()},
	}
	for _, mediaID := range s.MediaIDs {
		f.Add("media_ids", mediaID)
	}
	if s.ReplyToID != "" {
		f.Set("in_reply_to_id", s.ReplyToID)
	}
	if s.Sensitive {
		f.Set("sensitive", "true")
	}
	if s.Spoiler != "" {
		f.Set("spoiler_text", s.Spoiler)
	}
	return f
}

type statusResponse struct {
	ID    string `json:"id"`
	Error string `json:"error,omitempty"`
}

func (s Status) Submit(ctx context.Context, instance, appName string) (id string, err error) {
	var accessToken string
	accessToken, err = app.GetAccessToken(ctx, instance, appName)
	if err != nil {
		return
	}

	u := fmt.Sprintf("https://%s/api/v1/statuses", instance)
	f := s.FormData()

	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, u, strings.NewReader(f.Encode()))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	var respBody []byte
	respBody, err = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return
	}

	var response statusResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		err = fmt.Errorf("got status %d: %s %w", resp.StatusCode, string(respBody), err)
		return
	}
	if response.Error != "" {
		err = fmt.Errorf("got status %d: %s", resp.StatusCode, response.Error)
		return
	}

	id = response.ID
	return
}
