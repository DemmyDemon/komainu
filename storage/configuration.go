package storage

import "log"

// GetConfiguration gets a freshly loaded configuration.
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

// Path returns the path to where the configuration is stored.
func (c *Configuration) Path() string {
	return "data/config.json"
}

// Load loads the configuration file, or creates one if none exists.
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

// Save, in a shocking turn of events, saves the configuration file.
func (c *Configuration) Save() error {
	return SaveJSON(c)
}
