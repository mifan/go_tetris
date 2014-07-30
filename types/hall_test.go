package types

import (
	"encoding/json"
	"testing"
)

func Test_NormalHall(t *testing.T) {
	h := NewNormalHall()

	if err := h.NewTable(h.NextTableId(), "damon", "192.122.14.1", 0); err != nil {
		t.Error(err)
	}

	table := h.GetTableById(1)

	if table.GetIp() != "192.122.14.1" {
		t.Error("host not correct")
	}
	if table.Expire() {
		t.Error("should not expire so soon")
	}
	if table.IsGamble() {
		t.Error("it is not gamble")
	}
	if table.IsStart() {
		t.Error("it is not start")
	}

	table.Join(NewUser(1, "lol@163.com", "hi", "zoneT_T", "Lt1423"))
	table.Join(NewUser(2, "sbchao@qq.com", "hi", "sb_chao", "Lt1323"))
	if err := table.Join(NewUser(3, "schao@qq.com", "hi", "s_chao", "Lt1323")); err != nil {
		t.Logf("yeah you are right, the error should not be nil, but %v", err)
	}
	table.SwitchReady(1)
	table.SwitchReady(2)
	table.Quit(2)

	b, _ := h.MarshalJSON()
	t.Log(string(b))
}

func Test_TournamentHall(t *testing.T) {
	nh := NewNormalHall()
	th := NewTournamentHall(1<<4, 7, 3, "192.168.0.1:9901")

	b, _ := json.Marshal(map[string]interface{}{
		"normal":     nh,
		"tournament": th,
	})
	t.Log(string(b))
}
