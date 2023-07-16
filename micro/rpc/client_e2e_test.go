package rpc

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_InitClientProxy(t *testing.T) {
	addr := ":8081"
	server := NewServer()
	server.RegisterService(&UserServiceServer{})
	go func() {
		err := server.Start("tcp", addr)
		if err != nil {
			t.Log(err)
		}
	}()
	time.Sleep(3 * time.Second)
	usClient := &UserService{}
	err := InitClientProxy(addr, usClient)
	require.NoError(t, err)
	resp, err := usClient.GetByID(context.Background(), &GetByIDReq{ID: 123})
	require.NoError(t, err)
	assert.Equal(t, &GetByIDResp{
		Msg: fmt.Sprintf("recieved msg: %d", 123),
	}, resp)
}
