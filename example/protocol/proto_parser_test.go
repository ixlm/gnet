package protocol

import (
	"fmt"
	"testing"
)

func TestConvert(t *testing.T) {
	t.Run("NetHeader-encode_decode", func(t *testing.T) {
		header := NetHeader{
			PackLen:     10,
			MagicNumber: 20,
			MsgId:       30,
			Mark:        40,
			SessId:      50,
		}

		encodedBuf, _ := header.Bytes()

		newHeader, _ := NewHeaderFromBytes(encodedBuf)

		fmt.Printf("original header is %v\n", header)
		fmt.Printf("converted header is %v\n", newHeader)
		if !header.Equal(newHeader) {
			t.Fatal("convert failed")
		}

	})
}
