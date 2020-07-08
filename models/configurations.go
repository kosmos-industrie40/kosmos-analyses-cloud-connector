package models

// Password is the configuration which contains the user and password configurations
type Password struct {
	Mqtt struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"mqtt"`
	Database struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"database"`
}

// Configurations contains all other configuration details
type Configurations struct {
	Webserver struct {
		Address string `yaml:"address"`
		Port    int    `yaml:"port"`
	} `yaml:"webserver"`
	Database struct {
		Address  string `yaml:"address"`
		Port     int    `yaml:"port"`
		Database string `yaml:"database"`
	} `yaml:"database"`
	Mqtt struct {
		Address string `yaml:"address"`
		Port    int    `yaml:"port"`
	} `yaml:"mqtt"`
}
