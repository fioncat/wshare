package crypto

import (
	"reflect"
	"testing"
)

func TestEncrypt(t *testing.T) {
	password := "test12345"
	err := Init(password)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("This is a secrect message!!!!!!")

	result := Encrypt(data)
	raw, err := Decrypt(result)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(raw, data) {
		t.Fatal("unexpect decrypt result")
	}

	err = Init("wrong password!")
	if err != nil {
		t.Fatal(err)
	}

	raw, err = Decrypt(result)
	if err == nil {
		t.Fatal("expect error")
	}
	if reflect.DeepEqual(raw, data) {
		t.Fatal("expect message not equal")
	}
}
