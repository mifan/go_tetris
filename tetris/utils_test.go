package tetris

import "testing"

type testUtilsDataStruct struct {
	isOverlapped, isContiguous bool
	d1, d2                     dot
}

var testUtilsData = []testUtilsDataStruct{
	testUtilsDataStruct{
		isContiguous: true,
		isOverlapped: false,
		d1:           newDot(0, 0, newColor(1)),
		d2:           newDot(1, 0, newColor(1)),
	},
	testUtilsDataStruct{
		isContiguous: false,
		isOverlapped: true,
		d1:           newDot(0, 0, newColor(1)),
		d2:           newDot(0, 0, newColor(1)),
	},
}

func Test_Utils(t *testing.T) {
	// test contiguous and overlap
	for _, v := range testUtilsData {
		if isContiguous(v.d1, v.d2) != v.isContiguous {
			t.Errorf("%v and %v should %v be contiguous", v.d1, v.d2, v.isContiguous)
		}
		if isOverlapped(v.d1, v.d2) != v.isOverlapped {
			t.Errorf("%v and %v should %v be overlapped", v.d1, v.d2, v.isOverlapped)
		}
	}
	// TODO: test coverage
}
