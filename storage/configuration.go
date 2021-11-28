package storage

import "log"

func GetConfiguration() Configuration {
	cfg := Configuration{}
	if err := cfg.Load(); err != nil {
		log.Fatalln("Failed to load configuration:", err)
	}
	return cfg
}

type Configuration struct {
	Logfile string
}

func (c *Configuration) Path() string {
	return "data/config.json"
}

func (c *Configuration) Load() error {
	if exist, err := JSONFileExists(c); err != nil {
		return err
	} else if exist {
		return LoadJSON(c)
	} else {
		log.Println("Configuration file not found, will create a new one!")
		c.Logfile = "komainu.log"
		return c.Save()
	}
}

func (c *Configuration) Save() error {
	return SaveJSON(c)
}
