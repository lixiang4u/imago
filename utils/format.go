package utils

import (
	"encoding/json"
	"fmt"
	"github.com/cespare/xxhash/v2"
)

func ToJsonString(v interface{}, pretty bool) string {
	var buf []byte
	if pretty {
		buf, _ = json.MarshalIndent(v, "", "\t")
	} else {
		buf, _ = json.Marshal(v)
	}
	return string(buf)
}

func HashString(data string) string {
	return fmt.Sprintf("%x", xxhash.Sum64String(data))
}

func CompressRate(rawSize, convertedSize int64) string {
	return fmt.Sprintf(`%.2f%%`, 100.0-float64(convertedSize)*100/float64(rawSize))
}
