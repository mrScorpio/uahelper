package configs

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Endpoint   string
	Bot        bool
	RdMd       bool
	StoreCycle int
	TrPort     string
	BotToken   string
	BotChat    string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("error loading .env file, use default config")
		return &Config{
			Endpoint:   "opc.tcp://localhost:62544",
			Bot:        false,
			RdMd:       false,
			StoreCycle: 666,
			TrPort:     ":22222",
			BotToken:   "",
			BotChat:    "",
		}
	}
	var bot, rdmd bool
	stcc := 66
	bot, err = strconv.ParseBool(os.Getenv("BOT"))
	if err != nil {
		bot = false
	}
	rdmd, err = strconv.ParseBool(os.Getenv("RD"))
	if err != nil {
		rdmd = false
	}
	stcc, err = strconv.Atoi(os.Getenv("STCC"))
	if err != nil {
		stcc = 666
	}
	trPort := os.Getenv("TRPORT")
	if trPort == "" {
		trPort = ":22222"
	}
	botToken := os.Getenv("BOTOKEN")
	botChat := os.Getenv("BOTCHAT")

	return &Config{
		Endpoint:   os.Getenv("EP"),
		Bot:        bot,
		RdMd:       rdmd,
		StoreCycle: stcc,
		TrPort:     trPort,
		BotToken:   botToken,
		BotChat:    botChat,
	}
}
