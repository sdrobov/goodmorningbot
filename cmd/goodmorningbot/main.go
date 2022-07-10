package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/sdrobov/goodmorningbot/internal"
	"github.com/sdrobov/goodmorningbot/internal/addons/cbrf"
	"github.com/sdrobov/goodmorningbot/internal/addons/fuckinggreatadvice"
	"github.com/sdrobov/goodmorningbot/internal/addons/today"
	"github.com/sdrobov/goodmorningbot/internal/addons/weather"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type config struct {
	AppID        int    `long:"app-id" description:"Telegram App ID" env:"TG_APP_ID" required:"true"`
	AppHash      string `long:"app-hash" description:"Telegram App hash" env:"TG_APP_HASH" required:"true"`
	Phone        string `long:"phone" description:"Phone" env:"TG_PHONE" required:"true"`
	ChatId       string `long:"chat-id" description:"Telegram chat id to write to" env:"TG_CHAT_ID" required:"true"`
	Schedule     string `long:"schedule" description:"Cron-format schedule" env:"SCHEDULE" required:"true"`
	BaseGreeting string `long:"base-greeting" description:"Base greeting" env:"BASE_GREETING" required:"true"`
	Cataas       string `long:"cataas-endpoint" description:"cataas.com endpoint" env:"CATAAS_ENDPOINT" required:"true"`
	SessionFile  string `long:"session-file" description:"File to store tg session" env:"SESSION_FILE"`

	FuckingGreatAdviceEndpoint string  `long:"fga-endpoint" description:"fucking-great-advice.ru API endpoint" env:"FGA_ENDPOINT"`
	IsDayOffEndpoint           string  `long:"ido-endpoint" description:"isdayoff.ru API endpoint" env:"IDO_ENDPOINT"`
	OpenWeatherMapApiKey       string  `long:"owm-api-key" description:"openweathermap.org API Key" env:"OWM_API_KEY"`
	OpenWeatherMapLatitude     float64 `long:"owm-lat" description:"openweathermap.org Latitude" env:"OWM_LAT"`
	OpenWeatherMapLongitude    float64 `long:"owm-lon" description:"openweathermap.org Longitude" env:"OWM_LON"`
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

func main() {
	var cfg config
	parser := flags.NewParser(&cfg, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	cr := cron.New()
	ctx, cancelFunc := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancelFunc()
	go func() {
		<-ctx.Done()
		log.Default().Print("received exit signal! Initialize graceful shutdown")
		cr.Stop()
	}()

	uselessAddons := []internal.UselessAddon{
		today.NewToday(cfg.IsDayOffEndpoint),
		fuckinggreatadvice.NewFuckingGreatAdvice(cfg.FuckingGreatAdviceEndpoint),
		weather.NewWeather(cfg.OpenWeatherMapApiKey, cfg.OpenWeatherMapLatitude, cfg.OpenWeatherMapLongitude),
		cbrf.NewUsdRate(),
		cbrf.NewEurRate(),
	}

	prompt := codePrompt{}

	flow := auth.NewFlow(
		auth.CodeOnly(cfg.Phone, prompt),
		auth.SendCodeOptions{},
	)

	opts := telegram.Options{}
	if cfg.SessionFile != "" {
		opts.SessionStorage = &session.FileStorage{
			Path: cfg.SessionFile,
		}
	}

	client := telegram.NewClient(
		cfg.AppID,
		cfg.AppHash,
		opts,
	)

	if err := client.Run(ctx, func(ctx context.Context) error {
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return errors.Wrap(err, "error authenticating in telegram")
		}

		s := message.NewSender(client.API())
		var builder *message.RequestBuilder
		chatId, err := strconv.ParseInt(cfg.ChatId, 10, 64)
		if err == nil {
			m := new(peers.Options).Build(client.API())
			c, err := m.ResolveChatID(ctx, chatId)
			if err != nil {
				log.Fatalf("can't resolve chat id %d: %v", chatId, err)
			}
			builder = s.To(c.InputPeer())
		} else {
			builder = s.Resolve(cfg.ChatId)
		}

		_, err = cr.AddFunc(cfg.Schedule, func() {
			msg := cfg.BaseGreeting
			for _, addon := range uselessAddons {
				newMsg, err := addon.GetMessage(msg)
				if err != nil {
					log.Default().Printf("error running addon %s: %v", addon.Name(), err)
				} else if newMsg == "" {
					log.Default().Printf("addon %s returned empty string", addon.Name())
				} else {
					msg = newMsg
				}
			}

			doc := message.PhotoExternal(
				cfg.Cataas+`&rand=`+strconv.FormatInt(time.Now().UnixMicro(), 10),
				html.String(nil, msg),
			)

			if _, err := builder.Media(ctx, doc); err != nil {
				log.Fatalf("send error: %v", err)
			}
		})
		if err != nil {
			return errors.Wrap(err, "can't add cron job")
		}
		cr.Run()

		return nil
	}); err != nil {
		panic(err)
	}
}
