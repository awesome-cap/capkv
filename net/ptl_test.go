package net

import "testing"

func TestDecodeAndEncode(t *testing.T) {
	args := []string{"set", "name", "nico"}
	data := encode(args)
	t.Log(data)
	args, _, err := decode(data)
	if err != nil{
		t.Fatal(err)
	}
	t.Log(args)
}


