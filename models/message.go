package models

type Message struct {
	Sender     string `json:"sender"`
	Receiver   string `json:"receiver"`
	Message    string `json:"message"`
	DateCrated int64  `json:"date_crated"`
	Read       int    `json:"read"`
}
