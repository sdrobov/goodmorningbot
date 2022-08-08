package today

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type Today struct {
	isDayOffEndpoint string
	log              *zap.SugaredLogger
}

func NewToday(isDayOffEndpoint string, log *zap.SugaredLogger) *Today {
	return &Today{isDayOffEndpoint: isDayOffEndpoint, log: log}
}

func (t *Today) Name() string {
	return "Today"
}

func (t *Today) GetMessage(oldMessage string) (string, error) {
	isDayOff := time.Now().Weekday() == time.Saturday || time.Now().Weekday() == time.Sunday
	if t.isDayOffEndpoint != "" {
		httpClient := http.Client{Timeout: 2 * time.Second}
		resp, err := httpClient.Get(t.isDayOffEndpoint)
		if err != nil {
			t.log.Errorw("error fetching if it is day off", zap.Error(err))
		} else {
			defer func(body io.ReadCloser) {
				_ = body.Close()
			}(resp.Body)

			idoResult, err := io.ReadAll(resp.Body)
			if err != nil {
				t.log.Errorw("error reading response if it is day off", zap.Error(err))
			} else {
				isDayOff = string(idoResult) == "1"
			}
		}
	}

	isDayOffMsg := "это рабочий день"
	if isDayOff {
		isDayOffMsg = "это не рабочий день"
	}

	weekDay := ""
	switch time.Now().Weekday() {
	case time.Monday:
		weekDay = "понедельник"
	case time.Tuesday:
		weekDay = "вторник"
	case time.Wednesday:
		weekDay = "среда"
	case time.Thursday:
		weekDay = "четверг"
	case time.Friday:
		weekDay = "пятница"
	case time.Saturday:
		weekDay = "суббота"
	case time.Sunday:
		weekDay = "воскресенье"
	}

	return fmt.Sprintf("%s Сегодня %s, %s", oldMessage, weekDay, isDayOffMsg), nil
}
