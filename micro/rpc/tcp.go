package rpc

import (
	"encoding/binary"
	"net"
)

func ReadMsg(conn net.Conn) ([]byte, error) {
	// 读数据长度
	lenBytes := make([]byte, numOfLengthBytes)
	_, err := conn.Read(lenBytes)
	if err != nil {
		return nil, err
	}

	// 数据有多长
	length := binary.BigEndian.Uint64(lenBytes)
	bs := make([]byte, length)
	_, err = conn.Read(bs)
	return bs, err
}

func EncodeMsg(data []byte) []byte {
	respLen := len(data)

	// 构造响应数据
	// res = respLen 的 64位表示 + data
	res := make([]byte, respLen+numOfLengthBytes)
	// 第一步:
	// 先把长度写进去前八个字节
	binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(respLen))
	// 第二步:
	// 写入数据
	copy(res[numOfLengthBytes:], data)
	return res
}
