package channel

import (
	"context"
)

// import (
// "go.uber.org/atomic"
// )

// 利用channel实现一个任务池

type Task func()

type TaskPool struct {
	tasks chan Task
	// close *atomic.Bool
	close chan struct{}
}

func NewTaskPool(numG int, capcity int) *TaskPool {
	tp := &TaskPool{
		tasks: make(chan Task, capcity),
		// close: atomic.NewBool(false),
		close: make(chan struct{}),
	}

	for i := 0; i < numG; i++ {
		go func() {
			// for t := range tp.tasks {
			// 	// TODO: 收到关闭信号后, 执行完剩余的任务
			// 	if tp.close.Load() {
			// 		return
			// 	}
			// 	t()
			// }

			for {
				select {
				case <-tp.close:
					return
				case task := <-tp.tasks:
					task()
				}
			}
		}()
	}

	return tp
}

// 提交任务
func (t *TaskPool) Submit(ctx context.Context, task Task) error {
	select {
	case t.tasks <- task:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// Close 方法会释放资源, 不要重复调用
func (t *TaskPool) Close() error {
	// t.close.Store(true)
	// t.close <- struct{}{} 只会有一个channel收到(收到就没了, 其他goroutine就收不到)

	// 而close之后, 每一个gorouting都会收到消息(相当于广播)
	// 注意重复调用close channel会出错(可以使用sync.Once)
	close(t.close)
	return nil
}
