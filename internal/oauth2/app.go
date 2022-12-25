package oauth2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const (
	oauthOOB = "urn:ietf:wg:oauth:2.0:oob"
)

func RegisterApp(ctx context.Context, instance, appName string) (err error) {
	var c *client
	c, err = newClient()
	if err != nil {
		return err
	}

	_ = c

	return nil
}

func GetAccessToken(ctx context.Context, instance, appName, username, password string) (err error) {
	return nil
}

//func example() {
//	app, err := registerApp(HostName, AppName)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	jar, err := getOAuthCookies(HostName, app.ClientID)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	location, err := signin(HostName, jar, Username, Password)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	code, err := getOAuthCode(HostName, location, jar)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	token, err := getOAuthToken(HostName, app.ClientID, app.ClientSecret, code)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	log.Println(token)
//}

type registerAppRequest struct {
	ClientName   string `json:"client_name"`
	RedirectURIs string `json:"redirect_uris"`
	Scopes       string `json:"scopes"`
}

type registerAppResponse struct {
	ID           string `json:"id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func registerApp(hostname, appName string) (result registerAppResponse, err error) {
	u := fmt.Sprintf("https://%s/api/v1/apps", hostname)

	reqBody := new(bytes.Buffer)
	err = json.NewEncoder(reqBody).Encode(registerAppRequest{
		ClientName:   appName,
		RedirectURIs: oauthOOB,
		Scopes:       "read write",
	})
	if err != nil {
		return result, err
	}

	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, u, reqBody)
	if err != nil {
		return result, err
	}

	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return result, err
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("got status %d", resp.StatusCode)
	}

	var respBody []byte
	respBody, err = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(respBody, &result)
	return result, err
}

func getOAuthCookies(hostname, clientID string) (jar http.CookieJar, err error) {
	q := make(url.Values)
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", oauthOOB)
	q.Set("scope", "read write")

	u := fmt.Sprintf("https://%s/oauth/authorize?%s", hostname, q.Encode())
	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusSeeOther {
		return nil, fmt.Errorf("got status %d", resp.StatusCode)
	}

	jar, err = cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	jar.SetCookies(req.URL, resp.Cookies())

	return jar, nil
}

func signin(hostname string, jar http.CookieJar, username, password string) (location string, err error) {
	u := fmt.Sprintf("https://%s/auth/sign_in", hostname)
	f := make(url.Values)
	f.Set("username", username)
	f.Set("password", password)

	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	var resp *http.Response
	resp, err = client.PostForm(u, f)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("got status %d", resp.StatusCode)
	}

	uu, _ := url.Parse(u)
	jar.SetCookies(uu, resp.Cookies())

	return resp.Header.Get("Location"), nil
}

func getOAuthCode(hostname, location string, jar http.CookieJar) (code string, err error) {
	u := fmt.Sprintf("https://%s%s", hostname, location)

	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	var resp *http.Response
	resp, err = client.Post(u, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("got status %d", resp.StatusCode)
	}

	var urn *url.URL
	urn, err = url.Parse(resp.Header.Get("Location"))
	if err != nil {
		return "", err
	}

	code = urn.Query().Get("code")
	return code, nil
}

type oauthTokenResponse struct {
	Token string `json:"access_token"`
	Type  string `json:"token_type"`
}

func getOAuthToken(hostname, clientID, clientSecret, code string) (token string, err error) {
	u := fmt.Sprintf("https://%s/oauth/token", hostname)
	f := make(url.Values)
	f.Set("grant_type", "authorization_code")
	f.Set("redirect_uri", oauthOOB)
	f.Set("scope", "read write")
	f.Set("client_id", clientID)
	f.Set("client_secret", clientSecret)
	f.Set("code", code)

	var resp *http.Response
	resp, err = http.DefaultClient.PostForm(u, f)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got status %d", resp.StatusCode)
	}

	var respBody []byte
	respBody, err = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return "", err
	}

	var tokenResp oauthTokenResponse
	err = json.Unmarshal(respBody, &tokenResp)
	if err != nil {
		return "", err
	}

	return tokenResp.Token, nil
}
