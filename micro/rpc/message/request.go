package message

import (
	"bytes"
	"encoding/binary"
)

const (
	spliter     = '\n'
	metaSpliter = '\r'
)

type Request struct {
	HeadLength uint32
	BodyLength uint32
	RequestID  uint32
	Version    uint8
	Compresser uint8
	Serializer uint8

	ServiceName string
	MethodName  string
	Meta        map[string]string
	Data        []byte
}

func EncodeReq(req *Request) []byte {
	// 总长度为 head 长度 + body 长度
	bs := make([]byte, req.HeadLength+req.BodyLength)

	// 1.写入头部长度
	binary.BigEndian.PutUint32(bs[:4], req.HeadLength)
	// 2.写入body长度
	binary.BigEndian.PutUint32(bs[4:8], req.BodyLength)
	// 3.写入 Request ID
	binary.BigEndian.PutUint32(bs[8:12], req.RequestID)
	// 4.写入 Version
	bs[12] = req.Version
	// 5.写入Compresser
	bs[13] = req.Compresser
	// 6.写入Serializer
	bs[14] = req.Serializer

	cur := bs[15:]
	// 7.写入ServiceName
	copy(cur, req.ServiceName)

	// 8.写入分割符(用于分割字符串数据)
	cur = cur[len(req.ServiceName):]
	cur[0] = spliter
	cur = cur[1:]

	// 9.写入MethodName
	copy(cur, req.MethodName)

	// 10.写入分割符(用于分割字符串数据)
	cur = cur[len(req.MethodName):]
	cur[0] = spliter
	cur = cur[1:]

	// 11.写入元数据
	for key, val := range req.Meta {
		copy(cur, key)
		cur = cur[len(key):]
		cur[0] = metaSpliter
		cur = cur[1:]
		copy(cur, val)
		cur = cur[len(val):]

		// 写入分割符(用于分割字符串数据) 如 \nkey1\rvalue1\nkey2\rvalue2\n
		cur[0] = spliter
		cur = cur[1:]
	}

	// 12.写入Data
	copy(cur, req.Data)

	return bs
}

func DecodeReq(data []byte) *Request {
	req := &Request{}
	// 1.解head长度
	req.HeadLength = binary.BigEndian.Uint32(data[:4])
	// 2.解body长度
	req.BodyLength = binary.BigEndian.Uint32(data[4:8])
	// 3.解Request ID
	req.RequestID = binary.BigEndian.Uint32(data[8:12])
	// 4.解Version
	req.Version = data[12]
	// 5.解Compresser
	req.Compresser = data[13]
	// 6.解Serializer
	req.Serializer = data[14]

	header := data[15:req.HeadLength] // 将 header 和 body 切割

	// 7.解ServiceName
	index := bytes.IndexByte(header, spliter)
	req.ServiceName = string(header[:index])

	header = header[index+1:]
	index = bytes.IndexByte(header, spliter)
	// 8.解MethodName
	req.MethodName = string(header[:index])
	header = header[index+1:]

	// 9.解原数据
	index = bytes.IndexByte(header, spliter)
	if index != -1 {
		meta := make(map[string]string)
		for index != -1 {
			pair := header[:index]
			pairIndex := bytes.IndexByte(pair, metaSpliter)
			key := string(pair[:pairIndex])
			val := string(pair[pairIndex+1:])
			meta[key] = val
			header = header[index+1:]
			index = bytes.IndexByte(header, spliter)
		}

		req.Meta = meta
	}

	if req.BodyLength > 0 {
		req.Data = data[req.HeadLength:]
	}

	return req
}

func (req *Request) CalculateHeaderLength() {
	length := 15 + len(req.ServiceName) + 1 + len(req.MethodName) + 1
	for key, val := range req.Meta {
		length += len(key)
		length++ // 分隔符
		length += len(val)
		length++ // 分隔符
	}
	req.HeadLength = uint32(length)
}

func (req *Request) CalculateBodyLength() {
	req.BodyLength = uint32(len(req.Data))
}
