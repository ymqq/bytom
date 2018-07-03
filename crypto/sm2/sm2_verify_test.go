package sm2

import (
	"encoding/hex"
	"testing"
)

func decodeString(s string) []byte {
	data, _ := hex.DecodeString(s)

	return data
}

func TestVerifyBytes(t *testing.T) {
	var tests = []struct {
		pubX []byte
		pubY []byte
		msg  []byte
		uid  []byte
		r    []byte
		s    []byte
	}{
		{
			pubX: decodeString("09f9df311e5421a150dd7d161e4bc5c672179fad1833fc076bb08ff356f35020"),
			pubY: decodeString("ccea490ce26775a52dc6ea718cc1aa600aed05fbf35e084a6632f6072da9ad13"),
			msg:  decodeString("6d65737361676520646967657374"),
			uid:  decodeString("31323334353637383132333435363738"),
			r:    decodeString("f5a03b0648d2c4630eeac513e1bb81a15944da3827d5b74143ac7eaceee720b3"),
			s:    decodeString("b1b6aa29df212fd8763182bc0d421ca1bb9038fd1f7f42d4840b69c485bbc1aa"),
		},
	}

	for i, tt := range tests {
		result := VerifyBytes(tt.pubX, tt.pubY, tt.msg, tt.uid, tt.r, tt.s)

		if !result {
			t.Errorf("result: %d is invalid!", i)
		}
	}
}

func TestSm2VerifyBytes(t *testing.T) {
	var tests = []struct {
		publicKey []byte
		hash      []byte
		signature []byte
	}{
		{
			publicKey: decodeString("04" + "09f9df311e5421a150dd7d161e4bc5c672179fad1833fc076bb08ff356f35020" + "ccea490ce26775a52dc6ea718cc1aa600aed05fbf35e084a6632f6072da9ad13"),
			hash:      decodeString("f0b43e94ba45accaace692ed534382eb17e6ab5a19ce7b31f4486fdfc0d28640"),
			signature: decodeString("f5a03b0648d2c4630eeac513e1bb81a15944da3827d5b74143ac7eaceee720b3" + "b1b6aa29df212fd8763182bc0d421ca1bb9038fd1f7f42d4840b69c485bbc1aa"),
		},
	}

	for i, tt := range tests {
		result := Sm2VerifyBytes(tt.publicKey, tt.hash, tt.signature)

		if !result {
			t.Errorf("result: %d is invalid!", i)
		}
	}
}