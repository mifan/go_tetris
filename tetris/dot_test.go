package tetris

import "testing"

func Test_Dot(t *testing.T) {
	d := newDot(0, 0, newColor(1))
	// test move left
	d1 := d.moveLeft()
	if !d1.isOverlapped(newDot(-1, 0, newColor(1))) {
		t.Error("should be overlap")
	}
	if !d.isContiguous(d1) {
		t.Error("should be contiguous")
	}

	t.Log(d)
	t.Log(d.rotate(newDot(5, 0, newColor(1))))
}
