package models

type Todo struct {
	Username string `json:"username,omitempty"`
	Task     string `json:"todo"`
	Complete bool   `json:"complete"`
	Id       string `json:"id"`
}
