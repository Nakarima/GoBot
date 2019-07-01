# GoBot
Simple telegram bot made using Go and [Telebot v2](https://godoc.org/gopkg.in/tucnak/telebot.v2).

## Installing
Simply put your bot token in main.go inside NewBot function. Then run in terminal `go run main.go bot.go` or `go build main.go bot.go` and then you can use binary file `./main`.

## Use in Telegram
You can add simple answers for specific text using `/pattern |text|:answer`. 
There are 4 options available for pattern: 
* |text| - find any occurence of "text" in messages.
* | text| - find any preffix or word "text" in messages.
* |text | - find any suffix or word "text" in messages.
* | text | - find any word "text" in messages.

You can delete any pattern using `/pattern |text|`.

You can set a reminder using `/schedules note|31.12.19 15:55` (dd:mm:yy HH:MM).
Bot will remind you 24 hours before given date.
All current reminders are available under `/schedules`. This also clears all old reminders.

All of patterns and schedules are located in .json files so you can safely relauch bot without data loss.