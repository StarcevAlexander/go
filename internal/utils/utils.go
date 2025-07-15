package utils

import "encoding/json"

func ToJSON(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
