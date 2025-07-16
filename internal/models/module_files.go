package models

type FileItem struct {
	Title    string `json:"title"`
	FileName string `json:"fileName"`
}

// FileGroup представляет группу файлов с ID
type FileGroup struct {
	ID    int        `json:"id"`
	Files []FileItem `json:"files"`
}

// FilesResponse представляет корневую структуру ответа
type FilesResponse struct {
	Files []FileGroup `json:"files"`
}
