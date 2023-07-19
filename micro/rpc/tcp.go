package rpc

import (
	"encoding/binary"
	"net"
)

func ReadMsg(conn net.Conn) ([]byte, error) {
	// 协议头 + 协议体

	lenBytes := make([]byte, numOfLengthBytes)
	_, err := conn.Read(lenBytes)
	if err != nil {
		return nil, err
	}

	headerLen := binary.BigEndian.Uint32(lenBytes[:4])
	bodyLen := binary.BigEndian.Uint32(lenBytes[4:])
	length := headerLen + bodyLen
	bs := make([]byte, length)
	_, err = conn.Read(bs[8:])
	copy(bs[:8], lenBytes)
	return bs, err
}
