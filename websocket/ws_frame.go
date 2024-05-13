package websocket

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
)

const FIN_MASK = 0b10000000
const MASKED_MASK = 0b10000000
const OPCODE_MASK = 0b00001111
const PAYLOAD_FIRST_BYTE_MASK = 0b01111111

const OPCODE_CLOSE = 0x8
const OPCODE_PING = 0x9
const OPCODE_PONG = 0xA

type WebSocketFrame struct {
	Fin        bool
	OpCode     byte
	Masked     bool
	MaskingKey []byte
	Payload    []byte
}

func (frame WebSocketFrame) String() string {
	var buffer bytes.Buffer

	fin, masked := 0, 0
	if frame.Fin {
		fin = 1
	}
	if frame.Masked {
		masked = 1
	}

	buffer.WriteString(fmt.Sprintf("FIN=%d\n", fin))
	buffer.WriteString(fmt.Sprintf("OPCODE=%x\n", frame.OpCode))
	buffer.WriteString(fmt.Sprintf("MASKED=%d\n", masked))
	buffer.WriteString(fmt.Sprintf("PAYLOAD_LEN=%d\n", len(frame.Payload)))
	buffer.WriteString(fmt.Sprintf("MASKING_KEY=%b\n", frame.MaskingKey))
	buffer.WriteString(fmt.Sprintf("PAYLOAD=%x\n", frame.Payload))

	return buffer.String()
}

func NewCloseFrame(statusCode uint16, reason string, masked bool) WebSocketFrame {
	var buffer bytes.Buffer

	buffer.Write(binary.BigEndian.AppendUint16([]byte{}, statusCode))
	buffer.WriteString(reason)

	return generateControlFrame(OPCODE_CLOSE, masked, buffer.Bytes())
}

func NewPingFrame(data []byte, masked bool) WebSocketFrame {
	return generateControlFrame(OPCODE_PING, masked, data)
}

func NewPongFrame(data []byte, masked bool) WebSocketFrame {
	return generateControlFrame(OPCODE_PONG, masked, data)
}

func generateControlFrame(opCode byte, masked bool, data []byte) WebSocketFrame {
	if len(data) > 125 {
		data = data[:124]
	}

	var maskingKey []byte
	if masked {
		maskingKey = generateMaskingKey()
	}

	return WebSocketFrame{
		Fin:        true,
		OpCode:     opCode,
		Masked:     masked,
		MaskingKey: maskingKey,
		Payload:    data,
	}
}

func NewTextFrame(masked bool, text string) WebSocketFrame {
	var maskingKey []byte
	if masked {
		maskingKey = generateMaskingKey()
	}

	return WebSocketFrame{
		Fin:        true,
		OpCode:     0x1,
		Masked:     masked,
		MaskingKey: maskingKey,
		Payload:    []byte(text),
	}
}

func NewBinaryFrame(masked bool, data []byte) WebSocketFrame {
	var maskingKey []byte
	if masked {
		maskingKey = generateMaskingKey()
	}

	return WebSocketFrame{
		Fin:        true,
		OpCode:     0x2,
		Masked:     masked,
		MaskingKey: maskingKey,
		Payload:    data,
	}
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

	maskingKey, err := deserializeMaskingKey(reader)

	wsFrame.MaskingKey = maskingKey

	payload, err := deserializePayload(reader, wsFrame.MaskingKey, payloadLength)

	if err != nil {
		return wsFrame, fmt.Errorf(
			"failed to deserialize payload: %w", err)
	}

	wsFrame.Payload = payload
	return wsFrame, nil
}

func deserializeFirstByte(reader *bufio.Reader) (bool, byte, error) {
	data, err := reader.ReadByte()
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

func (frame WebSocketFrame) Serialize() []byte {
	var buffer bytes.Buffer

	var firstByte byte = 0
	if frame.Fin {
		firstByte = FIN_MASK
	}

	firstByte |= frame.OpCode

	buffer.WriteByte(firstByte)

	payloadLength := generatePayloadLength(frame)
	if frame.Masked {
		payloadLength[0] |= MASKED_MASK
	}

	buffer.Write(payloadLength)

	if frame.Masked {
		buffer.Write(frame.MaskingKey)
	}

	for i, payloadByte := range frame.Payload {
		if frame.Masked {
			payloadByte = payloadByte ^ frame.MaskingKey[i%4]
		}
		buffer.WriteByte(payloadByte)
	}

	return buffer.Bytes()
}

func generatePayloadLength(frame WebSocketFrame) []byte {
	length := len(frame.Payload)

	if length <= 125 {
		return []byte{byte(length)}
	}

	if length < 635536 {
		buffer := []byte{126}
		return binary.BigEndian.AppendUint16(buffer, uint16(length))
	}

	buffer := []byte{127}
	return binary.BigEndian.AppendUint64(buffer, uint64(length))
}

func generateMaskingKey() []byte {
	maskingKey := make([]byte, 4)
	maskingKeyValue, _ := rand.Int(rand.Reader, big.NewInt(4294967296))
	maskingKeyValue.FillBytes(maskingKey)
	return maskingKey
}
