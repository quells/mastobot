package toot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/quells/mastobot/internal/app"
	"io"
	"net/http"
	"net/url"
)

type verifyCredentialsResponse struct {
	AccountID string `json:"id"`
}

func VerifyCredentials(ctx context.Context, instance, appName string) (accountID string, err error) {
	var accessToken string
	accessToken, err = app.GetAccessToken(ctx, instance, appName)
	if err != nil {
		return
	}

	u := fmt.Sprintf("https://%s/api/v1/accounts/verify_credentials", instance)

	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")
	req = req.WithContext(ctx)

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		err = fmt.Errorf("invalid token")
		return
	}

	var respBody []byte
	respBody, err = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return
	}

	var response verifyCredentialsResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		err = fmt.Errorf("got status %d: %s %w", resp.StatusCode, string(respBody), err)
		return
	}

	accountID = response.AccountID
	return
}

type ListStatuses struct {
	MaxID          string // return results older than this ID
	SinceID        string // return results newer than this ID
	MinID          string // return results immediately newer than this ID
	Limit          int    // defaults to 20, max 40
	OnlyMedia      bool
	ExcludeReplies bool
	ExcludeReblogs bool
	OnlyPinned     bool
	Tagged         string // filter for statuses using this hashtag
}

func (l ListStatuses) QueryParams() url.Values {
	v := make(url.Values)
	SetNonZero(&v, "max_id", l.MaxID)
	SetNonZero(&v, "since_id", l.SinceID)
	SetNonZero(&v, "min_id", l.MinID)
	SetNonZero(&v, "limit", l.Limit)
	SetNonZero(&v, "only_media", l.OnlyMedia)
	SetNonZero(&v, "exclude_replies", l.ExcludeReplies)
	SetNonZero(&v, "exclude_reblogs", l.ExcludeReblogs)
	SetNonZero(&v, "pinned", l.OnlyPinned)
	SetNonZero(&v, "tagged", l.Tagged)
	return v
}

// ForAccount ID ListStatuses matching parameters sorted newest to oldest.
func (l ListStatuses) ForAccount(ctx context.Context, instance, appName, accountID string) (statuses []Status, err error) {
	var accessToken string
	accessToken, err = app.GetAccessToken(ctx, instance, appName)
	if err != nil {
		return
	}

	u := fmt.Sprintf("https://%s/api/v1/accounts/%s/statuses?%s", instance, accountID, l.QueryParams().Encode())

	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")
	req = req.WithContext(ctx)

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		err = fmt.Errorf("invalid token")
		return
	}

	var respBody []byte
	respBody, err = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return
	}

	err = json.Unmarshal(respBody, &statuses)
	if err != nil {
		err = fmt.Errorf("got status %d: %s %w", resp.StatusCode, string(respBody), err)
		return
	}

	return
}
