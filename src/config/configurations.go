package config

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
	UserMgmt struct {
		UserMgmt      string `yaml:"userMgmt"`
		ServerAddress string `yaml:"serverAddress"`
	} `yaml:"userMgmt"`
}
