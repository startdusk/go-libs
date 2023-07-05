package channel

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func Test_Broker_Send(t *testing.T) {
	b := &Broker{}

	// 模拟发送者
	go func() {
		for {
			err := b.Send(Msg{Content: time.Now().String()})
			if err != nil {
				t.Log(err)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)
		b.Close()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("我是消费者: %d", i)
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			msgs, err := b.Subscribe(10)
			if err != nil {
				return
			}
			for msg := range msgs {
				fmt.Println(name, msg.Content)
			}
		}(name)
	}
	wg.Wait()
}
