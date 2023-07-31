package main

import (
	"github.com/startdusk/go-libs/micro/rpc"
	"github.com/startdusk/go-libs/micro/rpc/serialize/json"
	"github.com/startdusk/go-libs/micro/rpc/serialize/proto"
)

func main() {
	svr := rpc.NewServer()
	svr.RegisterService(&UserService{})
	svr.RegisterService(&UserServiceProto{})
	svr.RegisterSerializer(json.Serializer{})
	svr.RegisterSerializer(proto.Serializer{})
	if err := svr.Start(":8081"); err != nil {
		panic(err)
	}
}
