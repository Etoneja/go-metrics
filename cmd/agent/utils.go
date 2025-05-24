package main

import (
	"fmt"
	"strconv"
	"strings"
)

func anyToString(v any) string {
	switch v := v.(type) {
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case float64:
		return fmt.Sprintf("%f", v)
	case float32:
		return fmt.Sprintf("%f", v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func buildURL(endpoint string, parts ...string) string {
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}
	return endpoint + strings.Join(parts, "/")
}
