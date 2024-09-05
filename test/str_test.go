package test

import "testing"

func TestStr(t *testing.T) {
	a := "aaa"
	b := "bbb"

	a += a + ">>>" + b

	t.Log(a)

}
