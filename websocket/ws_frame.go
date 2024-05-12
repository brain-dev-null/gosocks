package websocket

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

const FIN_MASK = 0b10000000
const MASKED_MASK = 0b10000000
const OPCODE_MASK = 0b00001111
const PAYLOAD_FIRST_BYTE_MASK = 0b01111111

type WebSocketFrame struct {
	Fin           bool
	OpCode        byte
	Masked        bool
	PayloadLength uint64
	MaskingKey    []byte
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
	wsFrame.OpCode = opcode

	masked, payloadLength, err := deserializePayloadLength(reader)
	if err != nil {
		return wsFrame, fmt.Errorf(
			"failed to deserialize payload length: %w", err)
	}

	wsFrame.Masked = masked
	wsFrame.PayloadLength = payloadLength

	maskingKey, err := deserializeMaskingKey(reader)

	wsFrame.MaskingKey = maskingKey

	payload, err := deserializePayload(reader, wsFrame.MaskingKey, wsFrame.PayloadLength)
	if err != nil {
		return wsFrame, fmt.Errorf(
			"failed to deserialize payload: %w", err)
	}

	wsFrame.Payload = payload
	return wsFrame, nil
}

func deserializeFirstByte(reader *bufio.Reader) (bool, byte, error) {
	data, err := readNextByte(reader)
	if err != nil {
		return false, 0, err
	}

	fin := data&FIN_MASK == FIN_MASK
	opcode := data & OPCODE_MASK

	return fin, opcode, nil
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

func deserializeMaskingKey(reader *bufio.Reader) ([]byte, error) {
	data := []byte{}

	for i := 0; i < 4; i++ {
		readByte, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}

		data = append(data, readByte)
	}

	return data, nil
}

func deserializePayload(reader *bufio.Reader, maskingKey []byte, payloadLength uint64) ([]byte, error) {
	payload := []byte{}

	for i := 0; i < int(payloadLength); i++ {
		data, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize payload byte %d/%d: %w",
				i, payloadLength, err)
		}

		payload = append(payload, data^maskingKey[i%4])
	}

	return payload, nil
}

func readNextByte(reader *bufio.Reader) (byte, error) {
	for {
		nextByte, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				continue
			}
			return 0, err
		}
		return nextByte, nil
	}
}
