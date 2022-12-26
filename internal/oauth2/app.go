package oauth2

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/doug-martin/goqu/v9"
	"github.com/quells/mastobot/internal/dbcontext"
	"github.com/rs/zerolog/log"
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

func GetAccessToken(ctx context.Context, instance, appName, email, password string) (err error) {
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
		Select("client_id", "client_secret").
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

	var clientID, clientSecret string
	err = db.QueryRow(query, params...).Scan(&clientID, &clientSecret)
	if err != nil {
		return err
	}

	err = c.getOAuthCookies(instance, clientID)
	if err != nil {
		return err
	}

	var signinLocation string
	signinLocation, err = c.signin(instance, email, password)
	if err != nil {
		return err
	}
	log.Debug().Str("location", signinLocation).Msg("sign-in location")

	var code string
	code, err = c.getOAuthCode(instance, signinLocation)
	if err != nil {
		return err
	}
	log.Debug().Str("code", code).Msg("sign-in code")

	var token string
	token, err = c.getOAuthToken(instance, clientID, clientSecret, code)
	if err != nil {
		return err
	}
	log.Debug().Str("token", token).Msg("access token")

	var stmt string
	stmt, params, err = goqu.
		Update("apps").
		Set(goqu.Record{
			"access_token": token,
		}).
		Where(goqu.Ex{
			"instance": instance,
			"app_name": appName,
		}).
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
		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		log.Debug().Msg(string(respBody))
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

func (c *client) getOAuthCookies(instance, clientID string) (err error) {
	c.ignoreRedirects()

	q := make(url.Values)
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", oauthOOB)
	q.Set("scope", "read write")

	u := fmt.Sprintf("https://%s/oauth/authorize?%s", instance, q.Encode())
	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	var resp *http.Response
	resp, err = c.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusSeeOther {
		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		log.Debug().Msg(string(respBody))
		return fmt.Errorf("got status %d while getting oauth2 cookies", resp.StatusCode)
	}

	return nil
}

func (c *client) signin(instance string, email, password string) (location string, err error) {
	c.ignoreRedirects()

	u := fmt.Sprintf("https://%s/auth/sign_in", instance)
	f := make(url.Values)
	f.Set("username", email)
	f.Set("password", password)

	var resp *http.Response
	resp, err = c.PostForm(u, f)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusFound {
		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		log.Debug().Msg(string(respBody))
		return "", fmt.Errorf("got status %d while signing in", resp.StatusCode)
	}

	return resp.Header.Get("Location"), nil
}

func (c *client) getOAuthCode(instance, location string) (code string, err error) {
	c.ignoreRedirects()
	u := fmt.Sprintf("https://%s%s", instance, location)

	var resp *http.Response
	resp, err = c.Post(u, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusFound {
		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		log.Debug().Msg(string(respBody))
		return "", fmt.Errorf("got status %d while getting oauth2 code", resp.StatusCode)
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

func (c *client) getOAuthToken(instance, clientID, clientSecret, code string) (token string, err error) {
	c.followRedirects()

	u := fmt.Sprintf("https://%s/oauth/token", instance)
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
		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		log.Debug().Msg(string(respBody))
		return "", fmt.Errorf("got status %d while getting oauth2 token", resp.StatusCode)
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
