package server

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"
)

type TangConfig struct {
	URL        string `json:"url"`
	Thumbprint string `json:"thp,omitempty"`
}

func (cfg *TangConfig) IsValidThumbprint(ctx context.Context) (bool, error) {
	if cfg.URL == "" {
		return false, errors.New("url is required")
	}
	if cfg.Thumbprint != "" {
		url := cfg.URL + "/adv/" + cfg.Thumbprint
		if !strings.Contains(url, "://") {
			url = "http://" + url
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return false, err
		}

		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()
		if resp.StatusCode == 404 {
			return false, errors.New("key does not exist")
		}
	}
	return true, nil
}

type DataPayload struct {
	Data     []byte `json:"data"`
	NodeUuid string `json:"node_uuid"`
}
