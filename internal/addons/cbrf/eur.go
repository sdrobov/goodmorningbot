package cbrf

import (
	"fmt"
	"github.com/ivanglie/go-cbr-client"
	"github.com/pkg/errors"
	"time"
)

type EurRate struct {
}

func NewEurRate() *EurRate {
	return &EurRate{}
}

func (e *EurRate) Name() string {
	return "EurRate"
}

func (e *EurRate) GetMessage(oldMessage string) (string, error) {
	cbrClient := cbr.NewClient()

	eurRate, err := cbrClient.GetRate("EUR", time.Now())
	if err != nil {
		return "", errors.Wrap(err, "error fetching EUR rate from cbr")
	}

	return fmt.Sprintf("%s\n\n<b>Курс EUR:</b> %.2f &#8381;", oldMessage, eurRate), nil
}
