package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/startdusk/go-libs/micro/registry"
	clientV3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// 使用etcd实现租约机制

type Registry struct {
	c    *clientV3.Client
	sess *concurrency.Session
}

// 这里 设计不是很好，理论上应该是要传接口的，但接口的方法没有实例的多(redis这个包设计得不好)
func NewRegistry(c *clientV3.Client) (*Registry, error) {
	sess, err := concurrency.NewSession(c)
	if err != nil {
		return nil, err
	}
	return &Registry{
		c:    c,
		sess: sess,
	}, nil
}

func (r *Registry) Register(ctx context.Context, si registry.ServiceInstance) error {
	val, err := json.Marshal(si)
	if err != nil {
		return err
	}
	_, err = r.c.Put(ctx, r.instanceKey(si), string(val), clientV3.WithLease(r.sess.Lease()))
	return err
}

func (r *Registry) UnRegister(ctx context.Context, si registry.ServiceInstance) error {
	_, err := r.c.Delete(ctx, r.instanceKey(si))
	return err
}
func (r *Registry) ListServices(ctx context.Context, serviceName string) ([]registry.ServiceInstance, error) {
	return nil, nil
}
func (r *Registry) Subscribe(ctx context.Context, serviceName string) (<-chan registry.Event, error) {
	return nil, nil
}

func (r *Registry) Close() error {
	return r.sess.Close()
}

func (r *Registry) instanceKey(si registry.ServiceInstance) string {
	return fmt.Sprintf("/micro/%s/%s", si.Name, si.Address)
}
