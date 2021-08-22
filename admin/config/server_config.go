package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode/utf8"
)

var ServerConfig *Config
var I18n map[string]string
var I18nJson string

type Config struct {
	ServerPort         int    `json:"server.port"`
	ServerContextPath  string `json:"server.context-path"`
	DataSourceUrl      string `json:"datasource.url"`
	DataSourceUsername string `json:"datasource.username"`
	DataSourcePassword string `json:"datasource.password"`
	DataSourceDriver   string `json:"datasource.driver"`
}

func init() {
	initServerConfig()
	initI18n()
}

func initServerConfig() {
	file, err := os.Open("conf.json")
	if err != nil {
		log.Fatalln("Read server config error ", err)
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalln("Read server config error ", err)
	}
	file.Close()

	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		log.Fatalf("Unmarshal %v server config error %v \n", string(bytes), err)
	}
	ServerConfig = &config
	file.Close()
}

func initI18n() {
	file, err := os.Open("i18n")
	if err != nil {
		log.Fatalln("Read server config error ", err)
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalln("Read server config error ", err)
	}
	file.Close()
	file.Close()

	prop := map[string]string{}

	lines := string(bytes)
	for _, line := range strings.Split(lines, "\n") {
		if utf8.RuneCountInString(lines) <= 0 {
			continue
		}

		if strings.Index(line, "##") == 0 {
			continue
		}

		parts := strings.Split(line, "=")
		if len(parts) < 2 {
			continue
		}
		prop[parts[0]] = parts[1]
	}
	I18n = prop
	b, _ := json.Marshal(I18n)
	I18nJson = string(b)
}
