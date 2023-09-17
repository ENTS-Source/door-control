package amember

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var ApiKey string
var ApiRootUrl string

func allPages[R any](endpoint string, query url.Values) ([]R, error) {
	perPage := 1000
	query.Set("_count", strconv.Itoa(perPage))

	page := 0
	expectedRecords := 0
	records := make([]R, 0)

	for {
		query.Set("_page", strconv.Itoa(page))
		page++

		val, err := doRequest[map[string]any](endpoint, query)
		if err != nil {
			return nil, err
		}

		if page == 1 { // we incremented up above already
			if t, ok := val["_total"]; ok {
				if total, ok := t.(float64); ok {
					expectedRecords = int(total)
				} else {
					return nil, errors.New("error parsing total")
				}
			} else {
				return nil, errors.New("error reading total")
			}
		}

		for k, v := range val {
			if strings.HasPrefix(k, "_") {
				continue
			}

			b, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}

			var row R
			err = json.Unmarshal(b, &row)
			if err != nil {
				return nil, err
			}

			records = append(records, row)
		}

		if len(records) >= expectedRecords {
			break
		}
	}

	return records, nil
}

func doRequest[R any](endpoint string, query url.Values) (R, error) {
	query.Set("_key", ApiKey)

	var zero R
	pth, err := url.JoinPath(ApiRootUrl, endpoint)
	if err != nil {
		return zero, err
	}
	req, err := http.NewRequest(http.MethodGet, pth, nil)
	if err != nil {
		return zero, err
	}
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)
	if res != nil {
		defer func(body io.ReadCloser) {
			_ = body.Close()
		}(res.Body)
	}
	if err != nil {
		return zero, nil
	}

	if res.StatusCode == http.StatusOK {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return zero, err
		}

		var val R
		err = json.Unmarshal(b, &val)
		if err != nil {
			return zero, err
		}

		return val, nil
	} else {
		return zero, fmt.Errorf("http error %d %s", res.StatusCode, res.Status)
	}
}
