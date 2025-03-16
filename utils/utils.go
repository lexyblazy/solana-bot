package utils

import "encoding/json"

func ToString(v interface{}) string {
	b, _ := json.Marshal(v)

	return (string(b))
}

func Deserialize[T any](input string) T {
	var result T

	json.Unmarshal([]byte(input), &result)

	return result

}
