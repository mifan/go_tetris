package types

import "encoding/json"

var (
	tU = &User{}
)

// json.Marshaler
var (
	_ json.Marshaler = tU
	_ json.Marshaler = NewObs()
	_ json.Marshaler = newTable(0, "", "", 0)
	_ json.Marshaler = NewTables()
)
