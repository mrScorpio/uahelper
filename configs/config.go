package configs

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/mrscorpio/uahelper/pkg/tagdata"
)

type Config struct {
	Endpoint     string
	Bot          bool
	VkBot        bool
	RdMd         bool
	StoreCycle   int
	TrPort       string
	BotToken     string
	BotChat      string
	BossId       string
	VkToken      string
	VkChat       string
	VkBossId     string
	ViewerUrl    string
	UaUser       string
	UaPass       string
	BrPath       string
	NatsAddr     string
	ShowTags     map[int]bool
	ShowTagNames []string
	Mu           sync.RWMutex `json:"-"`
}

func LoadConfig() *Config {
	data, err := os.ReadFile("cfg.json")
	showTags := make(map[int]bool)
	if err == nil {
		cfg := new(Config)
		err := json.Unmarshal(data, &cfg)
		cfg.ShowTags = showTags
		if err == nil {
			return cfg
		}
	}
	showTagNames := make([]string, 0)
	err = godotenv.Load()
	if err != nil {
		log.Println("error loading .env file, use default config")
		return &Config{
			Endpoint:     "opc.tcp://localhost:62544",
			Bot:          false,
			RdMd:         false,
			StoreCycle:   666,
			TrPort:       ":22222",
			BotToken:     "",
			BotChat:      "",
			UaUser:       "",
			UaPass:       "",
			BrPath:       "c:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe",
			NatsAddr:     "localhost:4222",
			ShowTags:     showTags,
			ShowTagNames: showTagNames,
		}
	}
	var bot, vkbot, rdmd bool
	stcc := 66
	bot, err = strconv.ParseBool(os.Getenv("BOT"))
	if err != nil {
		bot = false
	}
	vkbot, err = strconv.ParseBool(os.Getenv("VKBOT"))
	if err != nil {
		vkbot = false
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
	vkToken := os.Getenv("VKTOKEN")
	vkChat := os.Getenv("VKCHAT")
	bossId := os.Getenv("BOSSID")
	vkBossId := os.Getenv("VKBOSSID")
	viewerUrl := os.Getenv("VIEWERURL")
	uaUser := os.Getenv("UAUSER")
	uaPass := os.Getenv("UAPASS")
	brPath := os.Getenv("BRPATH")
	natsAddr := os.Getenv("NATS")
	return &Config{
		Endpoint:     os.Getenv("EP"),
		Bot:          bot,
		VkBot:        vkbot,
		RdMd:         rdmd,
		StoreCycle:   stcc,
		TrPort:       trPort,
		BotToken:     botToken,
		BotChat:      botChat,
		BossId:       bossId,
		VkToken:      vkToken,
		VkChat:       vkChat,
		VkBossId:     vkBossId,
		ViewerUrl:    viewerUrl,
		UaUser:       uaUser,
		UaPass:       uaPass,
		BrPath:       brPath,
		NatsAddr:     natsAddr,
		ShowTags:     showTags,
		ShowTagNames: showTagNames,
	}
}

func (c *Config) WrFile() error {
	c.Mu.RLock()
	data, err := json.MarshalIndent(&c, "", "  ")
	c.Mu.RUnlock()
	if err != nil {
		return err
	}
	err = os.WriteFile("cfg.json", data, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) UpdTagMap(d *tagdata.AllTags) {
	d.Mu.RLock()
	defer d.Mu.RUnlock()
	for _, v := range c.ShowTagNames {
		c.Mu.Lock()
		for i, tag := range d.Tag {
			if v == tag.Name {
				c.ShowTags[i] = true
				break
			}
		}
		c.Mu.Unlock()
	}
}
