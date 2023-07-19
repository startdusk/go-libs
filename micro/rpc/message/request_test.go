package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EncodeDecodeReq(t *testing.T) {
	cases := []struct {
		name string
		req  *Request
	}{
		{
			name: "normal",
			req: &Request{
				RequestID:   123,
				Version:     12,
				Compresser:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetByID",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b":      "a",
				},
				Data: []byte("Hello world"),
			},
		},
		{
			name: "no meta",
			req: &Request{
				RequestID:   123,
				Version:     12,
				Compresser:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetByID",
				Data:        []byte("Hello world"),
			},
		},
		{
			name: "no data",
			req: &Request{
				RequestID:   123,
				Version:     12,
				Compresser:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetByID",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b":      "a",
				},
			},
		},
		{
			name: "no meta and data",
			req: &Request{
				RequestID:   123,
				Version:     12,
				Compresser:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetByID",
			},
		},
		{
			name: "data with \n",
			req: &Request{
				RequestID:   123,
				Version:     12,
				Compresser:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetByID",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b":      "a",
				},
				Data: []byte("Hello\nworld"),
			},
		},
		{
			name: "data with \r",
			req: &Request{
				RequestID:   123,
				Version:     12,
				Compresser:  13,
				Serializer:  14,
				ServiceName: "user-service",
				MethodName:  "GetByID",
				Meta: map[string]string{
					"trace-id": "123456",
					"a/b":      "a",
				},
				Data: []byte("Hello\rworld"),
			},
		},

		// 禁止开发者在框架的Meta里面加\n \r
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.req.CalculateHeaderLength()
			c.req.CalculateBodyLength()
			data := EncodeReq(c.req)
			decodeReq := DecodeReq(data)
			assert.Equal(t, c.req, decodeReq)
		})
	}
}
