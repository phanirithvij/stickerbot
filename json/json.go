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
	FileID          string `json:"file_id"`
	StickerName     string `json:"sticker_name"`
	SampleStickerID string `json:"sample_sticker_id"`
}

// SaveToJSON ...
func SaveToJSON(data *Data, dest string) error {
	file, err := jsonp.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, file, 0644)
	return err
}

// LoadFromJSON ...
func LoadFromJSON(src string) (*Data, error) {

	d := new(Data)
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return d, err
	}

	err = jsonp.Unmarshal(data, &d)
	if err != nil {
		return d, err
	}
	return d, err
}
