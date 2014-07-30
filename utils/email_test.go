package utils

import "testing"

func Test_Email(t *testing.T) {
	err := SendMail("test", "coin tetris", "admin@cointetris.com")
	t.Log(err)
}
