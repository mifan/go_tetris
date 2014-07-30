package utils

import "strings"

func GetIp(ctx interface{}) string {
	if realIp := getContext(ctx).Request.Header.Get("X-Real-Ip"); realIp != "" {
		return realIp
	}
	return strings.Split(getContext(ctx).Request.RemoteAddr, ":")[0]
}
