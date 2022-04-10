package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

//Config struct
type Config struct {
	Database         string
	DatabaseUser     string `yaml:"database_user"`
	DatabasePassword string `yaml:"database_password"`
	DatabaseSocket   string `yaml:"database_socket"`
	DatabaseLocation string `yaml:"database_location"`
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
		return config, errors.New("wrong config file of path")
	}
	var configMain map[string]Config

	if err = yaml.Unmarshal(cfg, &configMain); err != nil {
		return Config{}, err
	}

	if config, ok = configMain[env]; !ok {
		return Config{}, fmt.Errorf("wrong work environment %s", env)
	}
	return config, nil

}
