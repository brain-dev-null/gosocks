package websocket

import (
	"bufio"
	"encoding/binary"
	"fmt"
)

const FIN_MASK = 0b10000000
const MASKED_MASK = 0b10000000
const OPCODE_MASK = 0b00001111
const PAYLOAD_FIRST_BYTE_MASK = 0b01111111

type WebSocketFrame struct {
	Fin           bool
	OpCode        string
	Masked        bool
	PayloadLength uint64
	MaskingKey    int
	Payload       []byte
}

func DeserialzeWebSocketFrame(reader *bufio.Reader) (WebSocketFrame, error) {
	wsFrame := WebSocketFrame{}

	fin, opcode, err := deserializeFirstByte(reader)
	if err != nil {
		return wsFrame, fmt.Errorf(
			"failed to deserialize first byte: %w", err)
	}

	wsFrame.Fin = fin
	wsFrame.OpCode = fmt.Sprintf("%b", opcode)

	masked, payloadLength, err := deserializePayloadLength(reader)
	if err != nil {
		return wsFrame, fmt.Errorf(
			"failed to deserialize payload length: %w", err)
	}

	wsFrame.Masked = masked
	wsFrame.PayloadLength = payloadLength

	return wsFrame, nil
}

func deserializeFirstByte(reader *bufio.Reader) (bool, int, error) {
	data, err := reader.ReadByte()
	if err != nil {
		return false, -1, err
	}

	fin := data&FIN_MASK == FIN_MASK
	opcode := data & OPCODE_MASK

	return fin, int(opcode), nil
}

func deserializePayloadLength(reader *bufio.Reader) (bool, uint64, error) {
	firstByte, err := reader.ReadByte()
	if err != nil {
		return false, 0, fmt.Errorf(
			"failed to read first byte of payload length: %w", err)
	}

	masked := firstByte&MASKED_MASK == MASKED_MASK
	firstBytePayloadLength := firstByte & PAYLOAD_FIRST_BYTE_MASK

	bytesToRead := 0

	if firstBytePayloadLength <= 125 {
		return masked, uint64(firstBytePayloadLength), nil
	}

	if firstBytePayloadLength == 126 {
		bytesToRead = 2
	} else if firstBytePayloadLength == 127 {
		bytesToRead = 8
	} else {
		return false, 0, fmt.Errorf(
			"unexpected first byte value: %d", firstBytePayloadLength)
	}

	bytesArray := []byte{}

	for i := 0; i < bytesToRead; i++ {
		nextByte, err := reader.ReadByte()
		if err != nil {
			return false, 0, fmt.Errorf(
				"failed to read byte %d/%d of payload length: %w",
				i, bytesToRead, err)
		}
		bytesArray = append(bytesArray, nextByte)
	}

	if bytesToRead == 2 {
		return masked, uint64(binary.BigEndian.Uint16(bytesArray)), nil
	} else if bytesToRead == 8 {
		return masked, uint64(binary.BigEndian.Uint64(bytesArray)), nil
	}

	return false, 0, fmt.Errorf(
		"invalid bytesToRead: %d", bytesToRead)
}
