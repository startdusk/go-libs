package main

import (
	"fmt"
	"github.com/startdusk/go-libs/micro"
	"github.com/startdusk/go-libs/micro/example/proto/gen"
	"github.com/startdusk/go-libs/micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func main() {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	if err != nil {
		panic(err)
	}
	r, err := etcd.NewRegistry(etcdClient)
	if err != nil {
		panic(err)
	}
	server := micro.NewServer("user-service", micro.ServerWithRegistry(r), micro.ServerWithTimeout(time.Second*3))
	us := &UserService{}
	// 我们将 UserService 什么样才算是初始化好的问题交给用户自己解决
	// 用户必须要在确认好 UserService 已经完全准备好之后才能启动并且注册
	gen.RegisterUserServiceServer(server, us)
	fmt.Println("启动服务器")
	if err = server.Start(":8081"); err != nil {
		fmt.Println(err)
	}
	server.Close()
}
