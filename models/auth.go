package models

type User struct {
	Name     string `json:"user"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

type RetUser struct {
	Name string `json:"user"`
}
