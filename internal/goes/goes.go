package goes

import (
	"bytes"
	"context"
	"github.com/nfnt/resize"
	"image/jpeg"
	"net/http"
)

/*
GOES-17

https://www.star.nesdis.noaa.gov/goes/index.php

256 thumb
1280 full

NOAA/NESDIS/STAR - GOES-West - GeoColor Composite by CIRA/NOAA
*/

const (
	goes17 = "https://cdn.star.nesdis.noaa.gov/GOES17/ABI/FD/GEOCOLOR/5424x5424.jpg"
)

func GOES17(ctx context.Context) (large, thumbnail []byte, err error) {
	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, goes17, nil)
	if err != nil {
		return
	}
	req = req.WithContext(ctx)

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	img, err := jpeg.Decode(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return
	}

	buf := new(bytes.Buffer)
	{
		m := resize.Resize(1280, 0, img, resize.Lanczos3)
		err = jpeg.Encode(buf, m, nil)
		if err != nil {
			return
		}
		large = buf.Bytes()
	}
	{
		m := resize.Resize(256, 0, img, resize.Lanczos3)
		err = jpeg.Encode(buf, m, nil)
		if err != nil {
			return
		}
		thumbnail = buf.Bytes()
	}

	return
}
