package configs

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Endpoint   string
	WrMd       bool
	RdMd       bool
	StoreCycle int
	TrPort     string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("error loading .env file, use default config")
		return &Config{
			Endpoint:   "opc.tcp://localhost:62544",
			WrMd:       true,
			RdMd:       false,
			StoreCycle: 666,
			TrPort:     ":22222",
		}
	}
	var wrmd, rdmd bool
	stcc := 66
	wrmd, err = strconv.ParseBool(os.Getenv("WR"))
	if err != nil {
		wrmd = true
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

	return &Config{
		Endpoint:   os.Getenv("EP"),
		WrMd:       wrmd,
		RdMd:       rdmd,
		StoreCycle: stcc,
		TrPort:     trPort,
	}
}
