package config

import (
	"errors"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

//Config struct
type Config struct {
	Database         string
	DatabaseUser     string `yaml:"database_user"`
	DatabasePassword string `yaml:"database_password"`
	DatabaseSocket   string `yaml:"database_socket"`
	Socket           string
	Domain           string
	Host             string
	ImagesPath       string `yaml:"images_path"`
	Logfile          string
	BotsList         string `yaml:"bots"`
}

//NewConfig config constructor
func NewConfig(configPath, env string) (Config, error) {

	var (
		ok     bool
		config Config
	)
	cfg, err := ioutil.ReadFile(configPath)

	if err != nil {
		return config, errors.New("Wrong config file of path")
	}
	var configMain map[string]Config

	err = yaml.Unmarshal(cfg, &configMain)

	if err != nil {
		log.Fatal(err)
	}

	if config, ok = configMain[env]; !ok {
		return Config{}, errors.New("Wrong work environment")
	}
	return config, nil

}
