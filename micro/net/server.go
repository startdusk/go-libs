package net

import (
	"errors"
	// "io"
	"net"
)

func Serve(network string, addr string) error {
	lis, err := net.Listen(network, addr)
	if err != nil {
		// 比较常见的错误就是端口被占用
		return err
	}

	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}

		go func() {
			if err := handleConn(conn); err != nil {
				_ = conn.Close()
			}
		}()
	}
}

func handleConn(conn net.Conn) error {
	for {
		// 读数据
		bs := make([]byte, 8)
		n, err := conn.Read(bs)
		// if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || errors.Is(err, io.ErrUnexpectedEOF) {
		// 	// 一般的关闭错误不用怎么管
		// 	// 也可以把日志输出
		// 	return err
		// }
		if err != nil {
			return err
		}
		// tcp一般不会有这个问题，但udp会有
		if n != 8 {
			return errors.New("micro: 没有读够数据")
		}

		res := handleMsg(bs)
		n, err = conn.Write(res)
		// if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || errors.Is(err, io.ErrUnexpectedEOF) {
		// 	// 一般的关闭错误不用怎么管
		// 	// 也可以把日志输出
		// 	return err
		// }
		if err != nil {
			return err
		}
		if n != len(res) {
			return errors.New("micro: 没写完数据")
		}
	}
}

func handleMsg(req []byte) []byte {
	res := make([]byte, 2*len(req))
	copy(res[:len(req)], req)
	copy(res[len(req):], req)
	return res
}
