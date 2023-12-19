package utils

import (
	"encoding/json"
	"fmt"
	"github.com/cespare/xxhash/v2"
)

func ToJsonString(v interface{}) string {
	bs, _ := json.MarshalIndent(v, "", "\t")
	return string(bs)
}

func HashString(data string) string {
	return fmt.Sprintf("%x", xxhash.Sum64String(data))
}

func CompressRate(rawSize, convertedSize int64) string {
	return fmt.Sprintf(`%.2f`, float64(convertedSize)/float64(rawSize))
}
