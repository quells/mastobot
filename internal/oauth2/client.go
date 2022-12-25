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

func (c *client) followingRedirects() *client {
	return &client{
		Client: http.Client{
			Jar: c.Client.Jar,
		},
	}
}

func ignoreRedirects(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}

func (c *client) ignoringRedirects() *client {
	return &client{
		Client: http.Client{
			Jar:           c.Client.Jar,
			CheckRedirect: ignoreRedirects,
		},
	}
}
