# Good Morning Bot

Бот, который не забудет пожелать "Доброе утро" коллегам и друзьям, а так же поднимет настроение
в начале дня картинкой с котиком

## Как запустить

Для запуска боту необходимы следующие параметры:
- флаг `--app-id` или env `TG_APP_ID` Telegram App ID
- флаг `--app-hash` или env `TG_APP_HASH` Telegram App hash
- флаг `--phone` или env `TG_PHONE` Телефон пользователя, от имени которого будем постить
- флаг `--chat-id` или env `TG_CHAT_ID` Юзернейм или ссылка на канал куда будем постить
- флаг `--schedule` или env `SCHEDULE` Расписание запуска в cron-формате
- флаг `--base-greeting` или env `BASE_GREETING` Начальная фраза-приветствие
- флаг `--cataas-endpoint` или env `CATAAS_ENDPOINT` Эндпоинт к cataas.com для получения котиков

Так же по желанию можно указать:
- флаг `--fga-endpoint` или env `FGA_ENDPOINT` Эндпоинт к api fucking-great-advice.ru - добавится "совет дня" (18+)
- флаг `--ido-endpoint` или env `IDO_ENDPOINT` Эндпоинт к api isdayoff.ru - так бот будет точнее определять нерабочие дни
- флаг `--owm-api-key` или env `OWM_API_KEY` API ключ к openweathermap.org для получения погоды
- флаг `--owm-lat` или env `OWM_LAT` Широту
- флаг `--owm-lon` или env `OWM_LON` Долготу


## Как дорабатывать

Очень просто: создайте новый модуль в internal/addons и в нем реализуйте интерфейс UselessAddon. Конструктор добавьте в
main.go в массив uselessAddons. Порядок следования элементов в массиве важен!
