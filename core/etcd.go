package core

import (
	"go.etcd.io/etcd/client/v3"
	"time"
)

func InitEtcd(add string) *clientv3.Client {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{add},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	return client
}
