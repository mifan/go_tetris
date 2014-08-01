package utils

import (
	"fmt"
	"reflect"
)

func HproseLog(fName string, params []reflect.Value, ctx interface{}) string {
	str := "invoker from " + GetIp(ctx) + "\n"
	str += "invokes the function " + fName + "\n"
	str += "parameters: "
	for _, v := range params {
		str += fmt.Sprintf("%v  ", v.Interface())
	}
	str += "\n"
	return str
}
