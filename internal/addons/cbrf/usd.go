package cbrf

import (
	"fmt"
	"github.com/ivanglie/go-cbr-client"
	"github.com/pkg/errors"
	"time"
)

type UsdRate struct {
}

func NewUsdRate() *UsdRate {
	return &UsdRate{}
}

func (u *UsdRate) Name() string {
	return "UsdRate"
}

func (u *UsdRate) GetMessage(oldMessage string) (string, error) {
	cbrClient := cbr.NewClient()

	usdRate, err := cbrClient.GetRate("USD", time.Now())
	if err != nil {
		return "", errors.Wrap(err, "error fetching USD rate from cbr")
	}

	return fmt.Sprintf("%s\n\n<b>Курс USD:</b> %.2f &#8381;", oldMessage, usdRate), nil
}
