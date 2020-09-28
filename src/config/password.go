package config

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
	UserMgmt struct {
		ClientId     string `yaml:"clientID"`
		ClientSecret string `yaml:"clientSecret"`
	} `yaml:"userMgmt"`
}
