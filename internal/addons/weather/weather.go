package weather

import (
	"fmt"
	owm "github.com/briandowns/openweathermap"
	"github.com/pkg/errors"
)

type Weather struct {
	apiKey string
	lat    float64
	lon    float64
}

func NewWeather(apiKey string, lat float64, lon float64) *Weather {
	return &Weather{apiKey: apiKey, lat: lat, lon: lon}
}

func (w *Weather) Name() string {
	return "Weather"
}

func (w *Weather) GetMessage(oldMessage string) (string, error) {
	if w.apiKey == "" {
		return oldMessage, nil
	}

	wc, err := owm.NewCurrent("C", "ru", w.apiKey)
	if err != nil {
		return "", errors.Wrap(err, "error creating owm client")
	}

	err = wc.CurrentByCoordinates(&owm.Coordinates{
		Longitude: w.lon,
		Latitude:  w.lat,
	})
	if err != nil {
		return "", errors.Wrap(err, "error fetching weather by coordinates")
	}

	return fmt.Sprintf(
		"%s\n\n<b>Погода в городе %s:</b> %.1f &#8451;\nОщущается как %.1f &#8451;",
		oldMessage,
		wc.Name,
		wc.Main.Temp,
		wc.Main.FeelsLike,
	), nil
}
