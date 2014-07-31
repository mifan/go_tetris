package utils

import "fmt"

// check if the config is empty value
func CheckEmptyConf(vals ...interface{}) {
forLoop:
	for _, v := range vals {
		switch v.(type) {
		case string:
			if v.(string) != "" {
				continue forLoop
			}
		case int, int32, int64, uint, uint32, uint64:
			if fmt.Sprintf("%v", v) != "0" {
				continue forLoop
			}
		}
		panic("something is wrong with the conf")
	}
}
