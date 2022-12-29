package toot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/quells/mastobot/internal/app"
)

type Visibility int

const (
	VisibilityInvalid  Visibility = iota
	VisibilityPrivate             // Visible only to followers and mentioned users, not on public timelines
	VisibilityUnlisted            // Visible to everyone, but does not appear on public timelines
	VisibilityPublic              // Visible to everyone and appears on public timelines
	VisibilityDirect              // Visible only to mentioned users
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

func (v *Visibility) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		return nil
	}
	switch len(data) {
	case 8:
		if bytes.Equal(data, []byte(`"public"`)) {
			*v = VisibilityPublic
			return nil
		}
		if bytes.Equal(data, []byte(`"direct"`)) {
			*v = VisibilityDirect
			return nil
		}
	case 9:
		if bytes.Equal(data, []byte(`"private"`)) {
			*v = VisibilityPrivate
			return nil
		}
	case 10:
		if bytes.Equal(data, []byte(`"unlisted"`)) {
			*v = VisibilityUnlisted
			return nil
		}
	}
	return fmt.Errorf("data is not a valid Visibility value")
}

type Status struct {
	ID         string     `json:"id"`
	Text       string     `json:"text"`
	MediaIDs   []string   `json:"media_ids"`
	ReplyToID  string     `json:"in_reply_to_id"`
	Sensitive  bool       `json:"sensitive"`
	Spoiler    string     `json:"spoiler_text"`
	Visibility Visibility `json:"visibility"`

	CreatedAt time.Time `json:"created_at"`
}

func (s Status) FormData() url.Values {
	f := url.Values{
		"status":     []string{s.Text},
		"visibility": []string{s.Visibility.String()},
	}
	SetNonZero(&f, "media_ids[]", s.MediaIDs)
	SetNonZero(&f, "in_reply_to_id", s.ReplyToID)
	SetNonZero(&f, "sensitive", s.Sensitive)
	SetNonZero(&f, "spoiler_text", s.Spoiler)
	return f
}

type statusResponse struct {
	ID    string `json:"id"`
	Error string `json:"error,omitempty"`
}

func (s Status) Submit(ctx context.Context, instance, appName string) (tootID string, err error) {
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

	tootID = response.ID
	return
}

func Delete(ctx context.Context, instance, appName, statusID string) (err error) {
	var accessToken string
	accessToken, err = app.GetAccessToken(ctx, instance, appName)
	if err != nil {
		return
	}

	u := fmt.Sprintf("https://%s/api/v1/statuses/%s", instance, statusID)

	var req *http.Request
	req, err = http.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got status %d", resp.StatusCode)
	}
	return nil
}
