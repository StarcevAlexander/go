package models

type Link struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type UserData struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Links []Link `json:"links"`
}
