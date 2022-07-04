package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/iArmann/cataas-API-go"
	"github.com/ivanglie/go-cbr-client"
	"github.com/jessevdk/go-flags"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type config struct {
	AppID   int    `long:"app-id" description:"Telegram App ID" env:"TG_APP_ID" required:"true"`
	AppHash string `long:"app-hash" description:"Telegram App hash" env:"TG_APP_HASH" required:"true"`
	Phone   string `long:"phone" description:"Phone" env:"TG_PHONE" required:"true"`
	ChatId  string `long:"chat-id" description:"Telegram chat id to write to" env:"TG_CHAT_ID" required:"true"`

	FuckingGreatAdviceEndpoint string `long:"fga-endpoint" description:"fucking-great-advice.ru API endpoint" env:"FGA_ENDPOINT"`
	IsDayOffEndpoint           string `long:"ido-endpoint" description:"isdayoff.ru API endpoint" env:"IDO_ENDPOINT"`
}

type codePrompt struct{}

func (p codePrompt) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

type fuckingGreatAdvice struct {
	Text string `json:"text,omitempty"`
}

func main() {
	var cfg config
	parser := flags.NewParser(&cfg, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	ctx, cancelFunc := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancelFunc()
	go func() {
		<-ctx.Done()
		log.Default().Print("received exit signal! Initialize graceful shutdown")
	}()

	client := telegram.NewClient(cfg.AppID, cfg.AppHash, telegram.Options{})

	if err := client.Run(context.Background(), func(ctx context.Context) error {
		prompt := codePrompt{}

		if err := auth.NewFlow(
			auth.CodeOnly(cfg.Phone, prompt),
			auth.SendCodeOptions{},
		).Run(ctx, client.Auth()); err != nil {
			log.Fatalf("error authenticating: %v", err)
		}

		cbrClient := cbr.NewClient()

		usdRate, err := cbrClient.GetRate("USD", time.Now())
		if err != nil {
			log.Fatalf("error fetching USD rate from cbr: %v", err)
		}

		eurRate, err := cbrClient.GetRate("EUR", time.Now())
		if err != nil {
			log.Fatalf("error fetching EUR rate from cbr: %v", err)
		}

		fgaResponse := &fuckingGreatAdvice{}
		if cfg.FuckingGreatAdviceEndpoint != "" {
			httpClient := http.Client{Timeout: 2 * time.Second}
			resp, err := httpClient.Get(cfg.FuckingGreatAdviceEndpoint)
			if err != nil {
				log.Fatalf("error fetching fucking great advice: %v", err)
			}

			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			jsonDecoder := json.NewDecoder(resp.Body)
			err = jsonDecoder.Decode(fgaResponse)
			if err != nil {
				log.Fatalf("error decoding fucking great advice response: %v", err)
			}
		}

		isDayOff := time.Now().Weekday() == time.Saturday || time.Now().Weekday() == time.Sunday
		if cfg.IsDayOffEndpoint != "" {
			httpClient := http.Client{Timeout: 2 * time.Second}
			resp, err := httpClient.Get(cfg.IsDayOffEndpoint)
			if err != nil {
				log.Fatalf("error fetching if it is day off: %v", err)
			}

			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			idoResult, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Fatalf("error reading response if it is day off: %v", err)
			}

			isDayOff = string(idoResult) == "1"
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

		msg := fmt.Sprintf("Доброе утро! Сегодня %s, %s", weekDay, isDayOffMsg)

		if fgaResponse.Text != "" {
			msg += fmt.Sprintf("\n\n<b>Совет дня:</b> %s", fgaResponse.Text)
		}

		msg += fmt.Sprintf("\n\n<b>Курс USD:</b> %.2f\n<b>Курс EUR:</b> %.2f", usdRate, eurRate)

		cats := &CataasAPI.Cataas{}
		cats.Size = CataasAPI.SIZE_SQUARE
		cats.Encode()
		data, err := cats.Get()
		if err != nil {
			log.Fatalf("can't get cats: %v", err)
		}

		u := uploader.NewUploader(client.API())
		s := message.NewSender(client.API()).WithUploader(u)
		up, err := u.FromBytes(ctx, "cat", data)
		if err != nil {
			log.Fatalf("can't upload cat: %v", err)
		}

		doc := message.UploadedDocument(up, html.String(nil, msg))
		doc.MIME("image/jpeg").Filename("cat.jpeg")

		target := s.Resolve(cfg.ChatId)
		if _, err := target.Media(ctx, doc); err != nil {
			log.Fatalf("send error: %v", err)
		}

		return nil
	}); err != nil {
		panic(err)
	}
}
