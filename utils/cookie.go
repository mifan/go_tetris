package utils

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/hprose/hprose-go/hprose"
	"github.com/xxtea/xxtea-go/xxtea"
)

// default domain
var domain = ".cointetris.com"

func SetDomain(val string) { domain = val }

// set a key to encrypt the cookie
var encKey []byte

// remember to set key before using utils
func SetCookieKey(key []byte) { encKey = key }

func getContext(ctx interface{}) *hprose.HttpContext { return ctx.(*hprose.HttpContext) }

// add cookie in the request
func AddCookie(name, val string, ctx interface{}) error {
	cookie := &http.Cookie{}
	cookie.Domain = domain
	cookie.Name = name
	cookie.MaxAge = 2 * 60 * 60 // 2 hours expire
	v, err := encrypt(val)
	if err != nil {
		return err
	}
	cookie.Value = v
	getContext(ctx).Request.AddCookie(cookie)
	return nil
}

// set cookie in response
func SetCookie(name, val string, ctx interface{}) error {
	cookie := &http.Cookie{}
	cookie.Domain = domain
	cookie.Name = name
	cookie.MaxAge = 2 * 60 * 60 // 2 hours expire
	v, err := encrypt(val)
	if err != nil {
		return err
	}
	cookie.Value = v
	http.SetCookie(getContext(ctx).Response, cookie)
	return nil
}

// get cookie
func GetCookie(name string, ctx interface{}) (cookie string, err error) {
	c, err := getContext(ctx).Request.Cookie(name)
	if err != nil {
		return
	}
	return decrypt(c.Value)
}

var errCanNotEncrypt = fmt.Errorf("can not encrypt")

// encrypt cookie
func encrypt(val string) (string, error) {
	data := xxtea.Encrypt([]byte(val), encKey)
	if data == nil {
		return "", errCanNotEncrypt
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

var errCanNotDecrypt = fmt.Errorf("can not decrypt")

// decrypt cookie
func decrypt(val string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return "", err
	}
	data := xxtea.Decrypt(b, encKey)
	if data == nil {
		return "", errCanNotDecrypt
	}
	return string(data), nil
}
