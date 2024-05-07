package websocket

import (
	"bufio"
	"bytes"
	"testing"
)

func TestFirstByteDeserialization(t *testing.T) {
	data := []byte{0b10001111, 0b10000001}
	reader := bufio.NewReader(bytes.NewReader(data))
	wsFrame, err := DeserialzeWebSocketFrame(reader)

	if err != nil {
		t.Fatalf("failed to deserialize ws frame: %v", err)
	}

	if !wsFrame.Fin {
		t.Fatalf("expected FIN to be 1, was 0")
	}

	if wsFrame.OpCode != "1111" {
		t.Fatalf("expected Opcode to be %s. got=%s", "1111", wsFrame.OpCode)
	}

	if wsFrame.PayloadLength != 1 {
		t.Fatalf("expected payload length to be %d. got=%d", 1, wsFrame.PayloadLength)
	}
}

func TestPayloadLengthDeserialization(t *testing.T) {
	tests := []struct {
		frameBytes            []byte
		expectedPayloadLength uint64
	}{
		{[]byte{0b10001111, 0b10000000}, 0},
		{[]byte{0b10001111, 0b10000001}, 1},
		{[]byte{0b10001111, 0b11111101}, 125},
		{[]byte{0b10001111, 0b11111110, 0b00000000, 0b01111110}, 126},
		{[]byte{0b10001111, 0b11111110, 0b00000000, 0b01111111}, 127},
		{[]byte{0b10001111, 0b11111110, 0b11111111, 0b11111111}, 65535},
		{[]byte{0b10001111, 0b11111111, 255, 255, 255, 255, 255, 255, 255, 255}, 18446744073709551615},
	}

	for _, tt := range tests {
		reader := bufio.NewReader(bytes.NewReader(tt.frameBytes))
		wsFrame, err := DeserialzeWebSocketFrame(reader)

		if err != nil {
			t.Errorf("failed to deserialize ws frame: %v", err)
			continue
		}

		if wsFrame.PayloadLength != tt.expectedPayloadLength {
			t.Errorf("expected payload length to be %d. got=%d",
				tt.expectedPayloadLength, wsFrame.PayloadLength)
		}
	}
}
