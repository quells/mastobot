package oauth2

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/quells/mastobot/internal/dbcontext"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const (
	oauthOOB   = "urn:ietf:wg:oauth:2.0:oob"
	appWebsite = "https://github.com/quells/mastobot"
)

func RegisterApp(ctx context.Context, instance, appName string) (err error) {
	var c *client
	c, err = newClient(ctx)
	if err != nil {
		return err
	}

	var db *sql.DB
	db, err = dbcontext.From(ctx)
	if err != nil {
		return err
	}

	var query string
	var params []any
	query, params, err = goqu.
		Select("instance").
		From("apps").
		Where(goqu.Ex{
			"instance": instance,
			"app_name": appName,
		}).
		ToSQL()
	if err != nil {
		return err
	}
	log.Debug().Msg(query)

	row := db.QueryRow(query, params...)

	var one string
	if qErr := row.Scan(&one); qErr != sql.ErrNoRows {
		err = fmt.Errorf("%q is already registered with %q", appName, instance)
		return err
	}

	var resp registerAppResponse
	resp, err = c.registerApp(instance, appName)
	if err != nil {
		return err
	}

	var stmt string
	stmt, params, err = goqu.
		Insert("apps").
		Cols("instance", "app_name", "app_id", "client_id", "client_secret").
		Vals(goqu.Vals{instance, appName, resp.AppID, resp.ClientID, resp.ClientSecret}).
		ToSQL()
	if err != nil {
		return err
	}
	log.Debug().Msg(stmt)

	_, err = db.ExecContext(ctx, stmt, params...)
	if err != nil {
		return err
	}

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

type registerAppResponse struct {
	AppID        string `json:"id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

func (c *client) registerApp(instance, appName string) (result registerAppResponse, err error) {
	c.followRedirects()
	u := fmt.Sprintf("https://%s/api/v1/apps", instance)
	f := make(url.Values)
	f.Set("client_name", appName)
	f.Set("redirect_uris", oauthOOB)
	f.Set("scopes", "read write")
	f.Set("website", appWebsite)

	var resp *http.Response
	resp, err = c.PostForm(u, f)
	if err != nil {
		return result, err
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("got status %d while registering app", resp.StatusCode)
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

func (c *client) getOAuthCookies(hostname, clientID string) (jar http.CookieJar, err error) {
	c.ignoreRedirects()

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

	var resp *http.Response
	resp, err = c.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusSeeOther {
		return nil, fmt.Errorf("got status %d while getting oauth2 cookies", resp.StatusCode)
	}

	jar, err = cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	jar.SetCookies(req.URL, resp.Cookies())

	return jar, nil
}

func (c *client) signin(hostname string, jar http.CookieJar, username, password string) (location string, err error) {
	c.ignoreRedirects()

	u := fmt.Sprintf("https://%s/auth/sign_in", hostname)
	f := make(url.Values)
	f.Set("username", username)
	f.Set("password", password)

	var resp *http.Response
	resp, err = c.PostForm(u, f)
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

func (c *client) getOAuthCode(hostname, location string, jar http.CookieJar) (code string, err error) {
	c.ignoreRedirects()
	u := fmt.Sprintf("https://%s%s", hostname, location)

	var resp *http.Response
	resp, err = c.Post(u, "application/x-www-form-urlencoded", nil)
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

func (c *client) getOAuthToken(hostname, clientID, clientSecret, code string) (token string, err error) {
	c.followRedirects()

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
