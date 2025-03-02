package utils

import "encoding/json"

func ToString(v interface{}) string {
	b, _ := json.Marshal(v)

	return (string(b))
}
