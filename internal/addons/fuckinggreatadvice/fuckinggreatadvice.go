package fuckinggreatadvice

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"time"
)

type fuckingGreatAdviceResponse struct {
	Text string `json:"text,omitempty"`
}

type FuckingGreatAdvice struct {
	endpoint string
}

func NewFuckingGreatAdvice(endpoint string) *FuckingGreatAdvice {
	return &FuckingGreatAdvice{endpoint: endpoint}
}

func (f *FuckingGreatAdvice) Name() string {
	return "FuckingGreatAdvice"
}

func (f *FuckingGreatAdvice) GetMessage(oldMessage string) (string, error) {
	if f.endpoint == "" {
		return oldMessage, nil
	}

	httpClient := http.Client{Timeout: 2 * time.Second}
	resp, err := httpClient.Get(f.endpoint)
	if err != nil {
		return "", errors.Wrap(err, "error fetching fucking great advice")
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	fgaResponse := &fuckingGreatAdviceResponse{}
	jsonDecoder := json.NewDecoder(resp.Body)
	err = jsonDecoder.Decode(fgaResponse)
	if err != nil {
		return "", errors.Wrap(err, "error decoding fucking great advice response")
	}

	return fmt.Sprintf("%s\n\n<b>Совет дня:</b> %s", oldMessage, fgaResponse.Text), nil
}
