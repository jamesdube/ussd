package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

func StringToSlice(s string) []string {
	return strings.Split(s, ".")
}

func Convert(b []byte, i interface{}) error {
	err := json.Unmarshal(b, &i)
	if err != nil {
		fmt.Println("Can;t unmarshal the byte array")
		return err
	}
	return nil
}
