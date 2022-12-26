package oauth2

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type client struct {
	http.Client
}

func newClient(ctx context.Context) (*client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	var timeout time.Duration
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	return &client{
		Client: http.Client{
			Jar:     jar,
			Timeout: timeout,
		},
	}, nil
}

func (c *client) followRedirects() {
	c.CheckRedirect = nil
}

func _ignoreRedirects(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

func (c *client) ignoreRedirects() {
	c.CheckRedirect = _ignoreRedirects
}
