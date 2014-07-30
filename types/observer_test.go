package types

import (
	"encoding/json"
	"log"
	"testing"
)

func Test_Observers(t *testing.T) {
	obs := NewObs()

	obs.Join(NewUser(1, "", "", "", ""))
	obs.Join(NewUser(2, "", "", "", ""))

	b, _ := json.Marshal(obs)
	log.Println(string(b))
}
