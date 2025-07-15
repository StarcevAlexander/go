package models

// Module представляет учебный модуль
type Module struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	DescriptionMin string `json:"descriptionMin"`
	DescriptionMax string `json:"descriptionMax"`
	TotalClasses   int    `json:"totalClasses"`
	TotalDuration  string `json:"totalDuration"`
	LinkToFolder   string `json:"linkToFolder"`
}
