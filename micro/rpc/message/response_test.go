package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EncodeDecodeResp(t *testing.T) {
	cases := []struct {
		name string
		resp *Response
	}{
		{
			name: "normal",
			resp: &Response{
				RequestID:  123,
				Version:    12,
				Compresser: 13,
				Serializer: 14,
				Error:      []byte("this is a error"),
				Data:       []byte("Hello world"),
			},
		},
		{
			name: "no error",
			resp: &Response{
				RequestID:  123,
				Version:    12,
				Compresser: 13,
				Serializer: 14,
				Data:       []byte("Hello world"),
			},
		},
		{
			name: "no data and error",
			resp: &Response{
				RequestID:  123,
				Version:    12,
				Compresser: 13,
				Serializer: 14,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.resp.CalculateHeaderLength()
			c.resp.CalculateBodyLength()
			data := EncodeResp(c.resp)
			decodeResp := DecodeResp(data)
			assert.Equal(t, c.resp, decodeResp)
		})
	}
}
