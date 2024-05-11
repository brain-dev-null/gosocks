package websocket

import (
	"bufio"
	"bytes"
	"testing"
)

func TestWebsocketFrameDeserialization(t *testing.T) {
	var FIN byte = 0b10000000
	var OPCODE byte = 2
	var MASKED byte = 128
	var MASKING_KEY_PATTERN byte = 0b10101010

	MASKING_KEY := repeatPattern(MASKING_KEY_PATTERN, 4)

	COMMON_HEADER := []byte{FIN + OPCODE}

	tests := []struct {
		data                 []byte
		encodedPayloadLength uint64
	}{
		{[]byte{MASKED + 0}, 0},
		{[]byte{MASKED + 1}, 1},
		{[]byte{MASKED + 125}, 125},
		{[]byte{MASKED + 126, 0, 126}, 126},
		{[]byte{MASKED + 126, 0, 127}, 127},
		{[]byte{MASKED + 126, 0, 128}, 128},
		{[]byte{MASKED + 127, 0, 0, 0, 0, 0, 1, 0, 0}, 65536},
	}

	for _, tt := range tests {
		PAYLOAD := repeatPattern(MASKING_KEY_PATTERN, tt.encodedPayloadLength)

		data := append(COMMON_HEADER, tt.data...)
		data = append(data, MASKING_KEY...)
		data = append(data, PAYLOAD...)

		reader := bufio.NewReader(bytes.NewReader(data))
		wsFrame, err := DeserialzeWebSocketFrame(reader)

		if err != nil {
			t.Errorf("failed to deserialize ws frame: %v", err)
			continue
		}

		if !wsFrame.Fin {
			t.Errorf("expected FIN to be 1, was 0")
			continue
		}

		if wsFrame.OpCode != OPCODE {
			t.Errorf("expected Opcode to be %d. got=%d", OPCODE, wsFrame.OpCode)
			continue
		}

		if wsFrame.PayloadLength != tt.encodedPayloadLength {
			t.Errorf("expected payload length to be %d. got=%d",
				tt.encodedPayloadLength, wsFrame.PayloadLength)
			continue
		}

		if len(wsFrame.MaskingKey) != 4 {
			t.Errorf("masking key has not length of 4. got=%d", len(wsFrame.MaskingKey))
			continue
		}

		for i, b := range MASKING_KEY {
			actualByte := wsFrame.MaskingKey[i]

			if actualByte != b {
				t.Errorf("bad masking key value at position %d. expected=%b, got=%b",
					i, b, actualByte)
			}
		}

		if wsFrame.PayloadLength != tt.encodedPayloadLength {
			t.Errorf("expected payload length to be %d. got=%d",
				tt.encodedPayloadLength, wsFrame.PayloadLength)
		}

		for i, b := range wsFrame.Payload {
			if b != 0 {
				t.Errorf("byte %d/%d is not 0. got=%d", i, len(wsFrame.Payload), b)
				continue
			}
		}
	}
}

func repeatPattern(pattern byte, repetitions uint64) []byte {
	data := []byte{}

	for i := uint64(0); i < repetitions; i++ {
		data = append(data, pattern)
	}

	return data
}
