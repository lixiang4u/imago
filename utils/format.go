package utils

import (
	"encoding/json"
	"fmt"
	"github.com/cespare/xxhash/v2"
	"github.com/gofiber/fiber/v2/utils"
	"reflect"
	"slices"
	"strings"
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
	return fmt.Sprintf(`%.2f`, 100.0-float64(convertedSize)*100/float64(rawSize))
}

func IsDefaultObj(obj interface{}, excludes []string) bool {
	v := reflect.ValueOf(obj)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		if slices.Contains(excludes, field.Name) {
			continue
		}
		if !IsDefaultValue(value.Interface()) {
			return false
		}
	}
	return true
}
func IsDefaultValue(value interface{}) bool {
	switch value := value.(type) {
	case string:
		return value == ""
	case int:
		return value == 0
	case float32:
		return value == 0
	case float64:
		return value == 0
	case bool:
		return value == false
	case []string:
		return len(value) == 0
	default:
		return false
	}
}

func UUIDv4() string {
	return utils.UUIDv4()
}

func FormattedUUID(length int) string {
	var s = strings.ReplaceAll(UUIDv4(), "-", "")
	return strings.ToLower(s[:length])
}

func FormatNickname(email string) string {
	var tmpList = strings.Split(email, "@")
	return tmpList[0]
}
