package toot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/quells/mastobot/internal/app"
)

type ContentTypeMedia string

const (
	ContentTypeMediaPNG  ContentTypeMedia = "image/png"
	ContentTypeMediaJPEG ContentTypeMedia = "image/jpeg"
)

type MediaUpload struct {
	ContentType ContentTypeMedia
	File        []byte
	Thumbnail   []byte
	Description string
	Focus       [2]float64
}

func (m MediaUpload) formatBody() (encoded []byte, contentType string, err error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)

	{
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				"file", "original"))
		h.Set("Content-Type", string(m.ContentType))
		var wi io.Writer
		wi, err = w.CreatePart(h)
		if err != nil {
			return
		}
		_, err = wi.Write(m.Thumbnail)
		if err != nil {
			return
		}
	}
	if len(m.Thumbnail) > 0 {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				"thumbnail", "thumbnail"))
		h.Set("Content-Type", string(m.ContentType))
		var wi io.Writer
		wi, err = w.CreatePart(h)
		if err != nil {
			return
		}
		_, err = wi.Write(m.Thumbnail)
		if err != nil {
			return
		}
	}
	if m.Description != "" {
		err = w.WriteField("description", m.Description)
		if err != nil {
			return
		}
	}
	if m.Focus[0] != 0 && m.Focus[1] != 0 {
		err = w.WriteField("focus", fmt.Sprintf("%.2f,%.2f", m.Focus[0], m.Focus[1]))
		if err != nil {
			return
		}
	}
	err = w.Close()
	if err != nil {
		return
	}

	contentType = w.FormDataContentType()
	encoded = buf.Bytes()
	return
}

type mediaUploadResponse struct {
	ID string `json:"id"`
}

func (m MediaUpload) Submit(ctx context.Context, instance, appName string) (mediaID string, err error) {
	var accessToken string
	accessToken, err = app.GetAccessToken(ctx, instance, appName)
	if err != nil {
		return
	}

	u := fmt.Sprintf("https://%s/api/v2/media", instance)

	var reqBody []byte
	var contentType string
	reqBody, contentType, err = m.formatBody()
	if err != nil {
		return
	}

	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, u, bytes.NewReader(reqBody))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", contentType)
	req = req.WithContext(ctx)

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

	statusRange := resp.StatusCode / 100
	if statusRange != 2 {
		err = fmt.Errorf("got status code %d: %s", resp.StatusCode, string(respBody))
		return
	}

	var uploadResp mediaUploadResponse
	err = json.Unmarshal(respBody, &uploadResp)
	if err != nil {
		return
	}

	mediaID = uploadResp.ID
	return
}
