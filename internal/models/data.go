package models

type Link struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type ModuleInfo struct {
	ModuleID int   `json:"module"`
	Date     int64 `json:"date"`
}

type UserData struct {
	ID      int          `json:"id"`
	Name    string       `json:"name"`
	Links   []Link       `json:"links"`
	Modules []ModuleInfo `json:"modules,omitempty"`
}
