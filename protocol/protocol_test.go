package protocol

import "testing"

func TestDecodeAndEncode(t *testing.T) {
	args := []string{"set", "name", "nico"}
	data := Encode(args)
	t.Log(data)
	args, _, err := Decode(data)
	if err != nil{
		t.Fatal(err)
	}
	t.Log(args)
}


