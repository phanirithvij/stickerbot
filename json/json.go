package json

import (
	jsonp "encoding/json"
	"io/ioutil"
)

// Data ..
type Data struct {
	Data []Pair `json:"data"`
}

// Pair ..
type Pair struct {
	FileID      string `json:"file_id"`
	StickerName string `json:"sticker_name"`
}

// SaveToJSON ...
func SaveToJSON(data *Data, dest string) error {
	file, _ := jsonp.MarshalIndent(data, "", " ")

	_ = ioutil.WriteFile("test.json", file, 0644)
	return nil
}
