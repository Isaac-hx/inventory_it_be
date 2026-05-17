// File this contain config file database,etc...
package config

import (
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

// Config struct type
type Config struct {

	//Application struct
	App struct {
		Port string `yaml:"port"`
	} `yaml:"app"`

	//Database struct
	Database struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Dbname   string `yaml:"db_name"`
	} `yaml:"database"`

	//JWT struct
	Jwt struct {
		Secretkey   string `yaml:"secret_key"`
		Expiredhour int    `yaml:"expired_hour"`
	}
	SuperAdmin struct {
		Username     string `yaml:"username"`
		Password     string `yaml:"password"`
		Email        string `yaml:"email"`
		DepartmentId string `yaml:"department_id"`
	}
}

// Load file .yaml config
func LoadConfig(path string) *Config {

	//Read file path
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("Cannot read config file", err)
	}

	//Instatiate config object
	var cfg Config

	//Parsing yaml file
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatal("Cannot parse config file", err)
	}

	//Return dereference object cfg
	return &cfg
}
