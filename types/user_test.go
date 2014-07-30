package types

import "testing"

func Test_Users(t *testing.T) {
	var sql string
	var args []interface{}
	us := NewUsers()

	us.SetNextId(1)

	us.Add(NewUser(us.GetNextId(), "hello@world.com", "password", "cointetris", "Ltcunixxx"))
	us.IncrNextId()

	if !us.IsEmailExist("hello@world.com") {
		t.Error("error, why not exist in ", us.emails)
	}

	us.Update(1, NewUpdateInt(UF_Balance, 1024))

	t.Log("here")
	t.Log(us.GetByEmail("hello@world.com"))
	t.Log("there")

	sql, args = us.GetByEmail("hello@world.com").SqlGeneratorUpdate()
	t.Log(sql)
	t.Log(args)
}
