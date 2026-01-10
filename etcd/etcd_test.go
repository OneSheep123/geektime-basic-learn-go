package etcd

import (
	"context"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// TestEtcdConnection 测试 etcd 连接
func TestEtcdConnection(t *testing.T) {
	// 创建 etcd 客户端配置
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:12379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("无法连接到 etcd: %v", err)
	}
	defer cli.Close()

	t.Log("成功连接到 etcd")

	// 测试基本的 Put 操作
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := "test-key"
	value := "test-value"

	_, err = cli.Put(ctx, key, value)
	if err != nil {
		t.Fatalf("Put 操作失败: %v", err)
	}
	t.Logf("成功写入键值对: %s = %s", key, value)

	// 测试 Get 操作
	resp, err := cli.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get 操作失败: %v", err)
	}

	if len(resp.Kvs) == 0 {
		t.Fatal("未找到键值对")
	}

	if string(resp.Kvs[0].Value) != value {
		t.Fatalf("期望值 %s, 实际值 %s", value, string(resp.Kvs[0].Value))
	}
	t.Logf("成功读取键值对: %s = %s", key, string(resp.Kvs[0].Value))

	// 测试 Delete 操作
	_, err = cli.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete 操作失败: %v", err)
	}
	t.Logf("成功删除键: %s", key)

	// 验证删除
	resp, err = cli.Get(ctx, key)
	if err != nil {
		t.Fatalf("验证删除时 Get 操作失败: %v", err)
	}

	if len(resp.Kvs) != 0 {
		t.Fatal("键应该已被删除")
	}
	t.Log("验证删除成功")
}

// TestEtcdWatch 测试 etcd Watch 功能
func TestEtcdWatch(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:12379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("无法连接到 etcd: %v", err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key := "watch-test-key"

	// 启动 watch
	watchChan := cli.Watch(ctx, key)

	// 在另一个 goroutine 中写入数据
	go func() {
		time.Sleep(1 * time.Second)
		_, err := cli.Put(ctx, key, "watch-value")
		if err != nil {
			t.Logf("Put 操作失败: %v", err)
		}
	}()

	// 等待 watch 事件
	select {
	case watchResp := <-watchChan:
		if len(watchResp.Events) == 0 {
			t.Fatal("未收到 watch 事件")
		}
		t.Logf("收到 watch 事件: %s = %s",
			string(watchResp.Events[0].Kv.Key),
			string(watchResp.Events[0].Kv.Value))
	case <-time.After(5 * time.Second):
		t.Fatal("等待 watch 事件超时")
	}

	// 清理
	_, err = cli.Delete(ctx, key)
	if err != nil {
		t.Logf("清理失败: %v", err)
	}
}

// TestEtcdLease 测试 etcd Lease 功能
func TestEtcdLease(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:12379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("无法连接到 etcd: %v", err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建一个 5 秒的租约
	leaseResp, err := cli.Grant(ctx, 5)
	if err != nil {
		t.Fatalf("创建租约失败: %v", err)
	}
	t.Logf("成功创建租约, ID: %d, TTL: %d", leaseResp.ID, leaseResp.TTL)

	key := "lease-test-key"
	value := "lease-value"

	// 使用租约写入数据
	_, err = cli.Put(ctx, key, value, clientv3.WithLease(leaseResp.ID))
	if err != nil {
		t.Fatalf("使用租约 Put 操作失败: %v", err)
	}
	t.Logf("成功使用租约写入键值对: %s = %s", key, value)

	// 验证数据存在
	resp, err := cli.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get 操作失败: %v", err)
	}
	if len(resp.Kvs) == 0 {
		t.Fatal("未找到键值对")
	}
	t.Log("验证数据存在")

	// 等待租约过期
	t.Log("等待租约过期...")
	time.Sleep(6 * time.Second)

	// 验证数据已被删除
	resp, err = cli.Get(ctx, key)
	if err != nil {
		t.Fatalf("验证过期时 Get 操作失败: %v", err)
	}
	if len(resp.Kvs) != 0 {
		t.Fatal("键应该已因租约过期而被删除")
	}
	t.Log("验证租约过期后数据已被删除")
}
