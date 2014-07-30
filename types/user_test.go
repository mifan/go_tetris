package types

import "testing"

func Test_Users(t *testing.T) {
	var sql string
	var args []interface{}
	us := NewUsers()

	us.SetNextId(1)

	us.Add(NewUser(us.GetNextId(), "pureveg@163.com", "password", "cointetris", "Ltcunixxx"))
	us.IncrNextId()

	if !us.IsEmailExist("pureveg@163.com") {
		t.Error("error, why not exist in ", us.emails)
	}

	us.Update(1, NewUpdateInt(UF_Balance, 1024))

	t.Log("here")
	t.Log(us.GetByEmail("pureveg@163.com"))
	t.Log("there")
	// us.Add(NewUser(us.GetNextId(), 1, 0, 0, 0, 0, "hello@qq.com", "pureveg", "Ltsss"))

	sql, args = us.GetByEmail("pureveg@163.com").SqlGeneratorUpdate()
	t.Log(sql)
	t.Log(args)
}
