package models

type UserData struct {
	ID      int          `json:"id"`
	Name    string       `json:"name"`
	Links   []Link       `json:"links"`
	Modules []ModuleInfo `json:"modules,omitempty"` // только у тьютора
}

type Link struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type ModuleInfo struct {
	Module int   `json:"module"` // ID модуля
	Date   int64 `json:"date"`   // дата окончания доступа
}
