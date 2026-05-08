package configs

import (
	"encoding/json"
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
	UaUser     string
	UaPass     string
	ShowTags   map[int]bool
}

func LoadConfig() *Config {
	data, err := os.ReadFile("cfg.json")
	if err == nil {
		cfg := new(Config)
		err := json.Unmarshal(data, &cfg)
		if err == nil {
			return cfg
		}
	}
	showTags := make(map[int]bool)
	err = godotenv.Load()
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
			UaUser:     "",
			UaPass:     "",
			ShowTags:   showTags,
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
	uaUser := os.Getenv("UAUSER")
	uaPass := os.Getenv("UAPASS")

	return &Config{
		Endpoint:   os.Getenv("EP"),
		Bot:        bot,
		RdMd:       rdmd,
		StoreCycle: stcc,
		TrPort:     trPort,
		BotToken:   botToken,
		BotChat:    botChat,
		UaUser:     uaUser,
		UaPass:     uaPass,
		ShowTags:   showTags,
	}
}

func (c *Config) WrFile() error {
	data, err := json.Marshal(&c)
	if err != nil {
		return err
	}
	err = os.WriteFile("cfg.json", data, 0755)
	if err != nil {
		return err
	}
	return nil
}
