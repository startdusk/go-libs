package net

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

type Pool struct {
	// 空闲连接队列
	idlesConns chan *idleConn
	// 请求队列
	reqQueue []connReq

	// 最大连接数
	maxCnt int

	// 当前连接数, 你已经建立好的连接
	cnt int

	// 最大空闲时间
	maxIdleTime time.Duration

	factory func() (net.Conn, error)

	lock sync.Mutex
}

func NewPool(
	initCnt int,
	maxIdleCnt int,
	maxCnt int,
	maxIdleTime time.Duration,
	factory func() (net.Conn, error)) (*Pool, error) {
	if initCnt > maxIdleCnt {
		return nil, errors.New("micro: 初始连接数量不能大于最大空闲数量")
	}
	idlesConns := make(chan *idleConn, maxIdleCnt)
	for i := 0; i < initCnt; i++ {
		conn, err := factory()
		if err != nil {
			return nil, err
		}
		idlesConns <- &idleConn{c: conn, lastActiveTime: time.Now()}
	}
	pool := &Pool{
		idlesConns:  idlesConns,
		maxCnt:      maxCnt,
		cnt:         0,
		maxIdleTime: maxIdleTime,
		factory:     factory,
	}

	return pool, nil
}

// 获取连接
func (p *Pool) Get(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	for {
		select {
		case ic := <-p.idlesConns:
			// 代表的是拿到了空闲连接
			if ic.lastActiveTime.Add(p.maxIdleTime).Before(time.Now()) {
				_ = ic.c.Close()
				continue
			}
			return ic.c, nil
		default:
			// 没有空闲连接
			p.lock.Lock()
			if p.cnt >= p.maxCnt {
				// 超过上限了
				req := connReq{connChan: make(chan net.Conn, 1)}
				p.reqQueue = append(p.reqQueue, req)
				p.lock.Unlock()
				select {
				case c := <-req.connChan: // 等别人归还
					return c, nil
				case <-ctx.Done():
					// 发生了超时
					// 选项1: 从队列里面删除掉 req 自己
					// 选项2: 在这里转发
					go func() {
						// 这里选择转发
						c := <-req.connChan
						_ = p.Put(context.Background(), c)
					}()
					return nil, ctx.Err()
				}
			}

			c, err := p.factory()
			if err != nil {
				p.lock.Unlock()
				return nil, err
			}
			p.cnt++
			p.lock.Unlock()
			return c, nil
		}
	}
}

// 归还连接
func (p *Pool) Put(ctx context.Context, c net.Conn) error {
	p.lock.Lock()
	if len(p.reqQueue) > 0 {
		// 有阻塞请求
		req := p.reqQueue[0]
		// 取走了第一个, 给它连接, 因此它获得了连接, 所以要把它从请求队列里面移除
		p.reqQueue = p.reqQueue[1:]
		p.lock.Unlock()
		req.connChan <- c
		return nil
	}

	p.lock.Unlock()
	// 没有阻塞请求
	ic := &idleConn{
		c:              c,
		lastActiveTime: time.Now(),
	}
	select {
	case p.idlesConns <- ic:
	default:
		// 空闲队列满了
		_ = c.Close()
		p.lock.Lock()
		p.cnt--
		p.lock.Unlock()
	}
	return nil
}

type idleConn struct {
	c              net.Conn
	lastActiveTime time.Time
}

type connReq struct {
	connChan chan net.Conn
}
