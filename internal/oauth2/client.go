package oauth2

import (
	"net/http"
	"net/http/cookiejar"
)

type client struct {
	http.Client
}

func newClient() (*client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &client{
		Client: http.Client{
			Jar: jar,
		},
	}, nil
}

func (c *client) followRedirects() {
	c.Client.CheckRedirect = nil
}

func ignoreRedirects(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

func (c *client) ignoreRedirects() {
	c.Client.CheckRedirect = ignoreRedirects
}
