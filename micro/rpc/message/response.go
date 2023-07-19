package message

import (
	"encoding/binary"
)

type Response struct {
	HeadLength uint32
	BodyLength uint32
	RequestID  uint32
	Version    uint8
	Compresser uint8
	Serializer uint8
	Error      []byte

	Data []byte
}

func EncodeResp(resp *Response) []byte {
	// 总长度为 head 长度 + body 长度
	bs := make([]byte, resp.HeadLength+resp.BodyLength)

	// 1.写入头部长度
	binary.BigEndian.PutUint32(bs[:4], resp.HeadLength)
	// 2.写入body长度
	binary.BigEndian.PutUint32(bs[4:8], resp.BodyLength)
	// 3.写入 Request ID
	binary.BigEndian.PutUint32(bs[8:12], resp.RequestID)
	// 4.写入 Version
	bs[12] = resp.Version
	// 5.写入 Compresser
	bs[13] = resp.Compresser
	// 6.写入 Serializer
	bs[14] = resp.Serializer

	cur := bs[15:]

	// 7.写入 Error
	copy(cur, resp.Error)

	cur = cur[len(resp.Error):]
	// 8.写入 Data
	copy(cur, resp.Data)

	return bs
}

func DecodeResp(data []byte) *Response {
	resp := &Response{}
	// 1.解head长度
	resp.HeadLength = binary.BigEndian.Uint32(data[:4])
	// 2.解body长度
	resp.BodyLength = binary.BigEndian.Uint32(data[4:8])
	// 3.解Request ID
	resp.RequestID = binary.BigEndian.Uint32(data[8:12])
	// 4.解Version
	resp.Version = data[12]
	// 5.解Compresser
	resp.Compresser = data[13]
	// 6.解Serializer
	resp.Serializer = data[14]

	// 7.解Error
	if resp.HeadLength > 15 {
		resp.Error = data[15:resp.HeadLength]
	}

	// 8.解Data
	if resp.BodyLength > 0 {
		resp.Data = data[resp.HeadLength:]
	}

	return resp
}

func (resp *Response) CalculateHeaderLength() {
	resp.HeadLength = 15 + uint32(len(resp.Error))
}

func (resp *Response) CalculateBodyLength() {
	resp.BodyLength = uint32(len(resp.Data))
}
