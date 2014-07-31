package utils

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/xxtea/xxtea-go/xxtea"
)

var tokenKey []byte

func SetTokenKey(key []byte) {
	tokenKey = key
}

const errCantGenerateToken = "can not generate token for user %s, %d"

func GenerateToken(uid int, nickname string, isApply bool, isOb bool, tid int) (string, error) {
	var token string
	// uid|nickname
	// uid|nickname|isObserver|tableId|isTournament
	if isApply {
		token = fmt.Sprintf("%d|%s", uid, nickname)
	} else {
		token = fmt.Sprintf("%d|%s|%v|%d|%v", uid, nickname, isOb, tid, tid >= 1e5)
	}
	b := xxtea.Encrypt([]byte(token), tokenKey)
	if b == nil {
		return "", fmt.Errorf(errCantGenerateToken, nickname, uid)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

const errTokenError = "the token %s is incorrect"

func ParseToken(token string) (uid int, nickname string, isApply bool, isOb bool, isTournament bool, tid int, err error) {
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return
	}
	token = string(xxtea.Decrypt(b, tokenKey))
	vals := strings.Split(token, "|")
	switch len(vals) {
	case 5:
		// join a normal game
		// uid
		uid, err = strconv.Atoi(vals[0])
		if err != nil {
			return
		}
		// nickname
		nickname = vals[1]
		// isOb
		isOb = vals[2] == "true"
		// tid
		tid, err = strconv.Atoi(vals[3])
		// isTournament
		isTournament = vals[4] == "true"
	case 2:
		// apply for tournament
		isApply = true
		// nickname
		nickname = vals[1]
		// uid
		uid, err = strconv.Atoi(vals[0])
	default:
		err = fmt.Errorf(errTokenError, token)
	}
	return
}
