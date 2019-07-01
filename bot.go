package main

import (
	"bytes"
	"encoding/json"
	tba "gopkg.in/tucnak/telebot.v2"
	"io/ioutil"
	"log"
	"strings"
	"text/template"
	"time"
)

type bot struct {
	api        *tba.Bot
	answersMap map[string]string
	schedules      []schedules
}

type schedules struct {
	Subject string
	Date    time.Time
	Chat    tba.Chat
}

const schedulesTemplate = `
Actual reminders:
{{ range . }}{{ .Subject }} - {{ .Date.Format "02.01.06 15:04"}}
{{ end }}
`

func clearText(text string) string {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, ".", "")
	text = strings.ReplaceAll(text, ",", "")
	text = strings.ReplaceAll(text, ";", "")
	text = strings.ReplaceAll(text, ":", "")
	text = strings.ReplaceAll(text, "!", "")
	var newText strings.Builder
	newText.WriteString(string(text[0]))
	for i := 1; i < len(text); i++ {
		if text[i] != text[i-1] {
			newText.WriteString(string(text[i]))
		}
	}
	return newText.String()
}

func patternAnswer(answersMap map[string]string, text string) string {
	for key, value := range answersMap {
		message := " " + clearText(text) + " "
		if strings.Contains(message, key) {
			return value
		}
	}
	return ""
}


func fastReply(bot *bot, message *tba.Message) {
	answer := patternAnswer(bot.answersMap, message.Text)
	if answer != "" {
		_, err := bot.api.Send(message.Chat, answer, &tba.SendOptions{
			ReplyTo: message,
		})
		if err != nil {
			log.Fatal("error answering")
		}
	}
}

func crudFastReply(bot *bot, message *tba.Message) {

		data := strings.ReplaceAll(message.Text, "/pattern ", "")
			if strings.Contains(data, ":") {
				data = strings.ReplaceAll(data, "|", "")
				splitData := strings.Split(data, ":")
				pattern := splitData[0]
				pattern = clearText(pattern)
				answer := splitData[1]
				bot.answersMap[pattern] = answer
				answers, _ := json.Marshal(bot.answersMap)
				err := ioutil.WriteFile("answers.json", answers, 0644)
				if err != nil {
					log.Fatal("error writing to answers file")
				} else {
					_, err = bot.api.Send(message.Chat, "added to patterns", &tba.SendOptions{
						ReplyTo: message,
					})
					if err != nil {
						log.Fatal("error sending message")
					}
				}
			} else {
				data = strings.ReplaceAll(data, "|", "")
				data = clearText(data)
				_, ok := bot.answersMap[data]
				if ok {
					delete(bot.answersMap, data)
					_, err := bot.api.Send(message.Chat, "pattern deleted", &tba.SendOptions{
						ReplyTo: message,
					})
					if err != nil {
						log.Fatal("error sending message")
					}
				}
				answers, _ := json.Marshal(bot.answersMap)
				err := ioutil.WriteFile("answers.json", answers, 0644)
				if err != nil {
					log.Fatal("error writing to answers file")
				}
			}
}

func checkScheduleExpiration(schedules []schedules) []schedules {
	for index, schedule := range schedules {
		if schedule.Date.Before(time.Now()) {
			schedules[index] = schedules[len(schedules)-1]
			schedules = schedules[:len(schedules)-1]
		}
	}
	schedulesJson, _ := json.Marshal(schedules)
	err := ioutil.WriteFile("schedules.json", schedulesJson, 0644)
	if err != nil {
		log.Fatal("error writing to answers file")
	}
	return schedules
}

func scheduleReminder(bot *bot, schedule schedules) {

	select {
	case <-time.After(schedule.Date.Sub(time.Now()) - 26*time.Hour):
		_, err := bot.api.Send(&schedule.Chat, "ACHTUNG!!1!! "+schedule.Subject+" - "+schedule.Date.Format("02.01.06 15:04"))
		if err != nil {
			log.Fatal("error sending reminder")
		}

	}
}

func scheduler(bot *bot, message *tba.Message) {
	if message.Text == "/schedules" {
		tmpl, err := template.New("schedulesTemplate").Parse(schedulesTemplate)
		if err != nil {
			log.Fatal("error parsing template")
		}
		bot.schedules = checkScheduleExpiration(bot.schedules)

		var answerBuff bytes.Buffer
		err = tmpl.Execute(&answerBuff, bot.schedules)
		if err != nil {
			log.Fatal("error executing template")
		}
		answer := answerBuff.String()

		_, err = bot.api.Send(message.Chat, answer, &tba.SendOptions{
			ReplyTo: message,
		})
		if err != nil {
			log.Fatal("error sending schedule")
		}
	} else {
		data := strings.ReplaceAll(message.Text, "/schedules ", "")
		if !strings.Contains(data, "/") {
			splitData := strings.Split(data, "|")
			subject := splitData[0]
			date := splitData[1]
			formattedDate, err := time.Parse("02.01.06 15:04", date)
			newExam := schedules{Subject: subject, Date: formattedDate, Chat: *message.Chat}
			bot.schedules = append(bot.schedules, newExam)
			go scheduleReminder(bot, newExam)
			schedules, _ := json.Marshal(bot.schedules)
			err = ioutil.WriteFile("schedules.json", schedules, 0644)
			if err != nil {
				log.Fatal("error writing to schedules file")
			} else {
				_, err = bot.api.Send(message.Chat, "Remind set", &tba.SendOptions{
					ReplyTo: message,
				})
				if err != nil {
					log.Fatal("error sending message")
				}
			}
		}
	}
}


func (bot *bot) Run() {
	for _, schedule := range bot.schedules {
		go scheduleReminder(bot, schedule)
	}
	bot.api.Handle("/version", func(m *tba.Message) {
		_, err := bot.api.Send(m.Chat, "v 1.0", &tba.SendOptions{
			ReplyTo: m,
		})
		if err != nil {
			log.Fatal("error checking version")
		}
	})

	bot.api.Handle("/schedules", func(m *tba.Message) {
		scheduler(bot, m)
	})

	bot.api.Handle("/pattern", func(m *tba.Message) {
		crudFastReply(bot, m)
	})

	bot.api.Handle(tba.OnText, func(m *tba.Message) {
		fastReply(bot, m)
	})

	bot.api.Start()
}

func NewBot(token string) *bot {
	tb, err := tba.NewBot(tba.Settings{
		Token:  token,
		Poller: &tba.LongPoller{Timeout: 1 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	answers := make(map[string]string)
	answersData, err := ioutil.ReadFile("answers.json")
	if err != nil {
		log.Fatal("error reading file")
	}

	err = json.Unmarshal([]byte(answersData), &answers)
	if err != nil {
		log.Fatal("error decoding json")
	}

	var schedules []schedules
	schedulesData, err := ioutil.ReadFile("schedules.json")
	if err != nil {
		log.Fatal("error reading file")
	}
	err = json.Unmarshal([]byte(schedulesData), &schedules)
	if err != nil {
		log.Fatal("error decoding json")
	}

	log.Printf("authorized")
	return &bot{tb, answers, schedules}
}
