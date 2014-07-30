package utils

import "testing"

func Test_Token(t *testing.T) {
	token, err := GenerateToken(10, "sbChao", true, false, 20)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("the token is ", token)
	}

	uid, nickname, isApply, isOb, isTournament, tid, err := ParseToken(token)
	if err != nil {
		t.Error(err)
	}

	if uid != 10 {
		t.Errorf("the uid should be %v", uid)
	}
	if nickname != "sbChao" {
		t.Errorf("the nicnname should not be %v", nickname)
	}
	if !isApply {
		t.Error("it should be applying for tournament")

		if tid != 20 {
			t.Errorf("the tid should be %v", tid)
		}
		if isOb {
			t.Error("it should be observing a game")
		}
	}
	if isTournament {
		t.Error("it should not be tournament")
	}
}
