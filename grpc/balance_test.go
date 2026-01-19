package grpc

import (
	"context"
	_ "ddd_demo/pkg/grpcx/balancer/wrr" // 自定义加权负载均衡算法
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/weightedroundrobin" //需要匿名引入基于权重的负载均衡算法
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type BalancerTestSuite struct {
	suite.Suite
	cli *etcdv3.Client
}

func (s *BalancerTestSuite) SetupSuite() {
	cli, err := etcdv3.NewFromURL("localhost:12379")
	// etcdv3.NewFromURLs()
	// etcdv3.New(etcdv3.Config{Endpoints: })
	require.NoError(s.T(), err)
	s.cli = cli
}

// failover 机制
func (s *BalancerTestSuite) TestFailoverClient() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`
{
  "loadBalancingConfig": [{"round_robin": {}}],
  "methodConfig":  [
    {
      "name": [{"service":  "UserService"}],
      "retryPolicy": {
        "maxAttempts": 4,
        "initialBackoff": "0.01s",
        "maxBackoff": "0.1s",
        "backoffMultiplier": 2.0,
        "retryableStatusCodes": ["UNAVAILABLE"]
      }
    }
  ]
}
`),
		//	"maxAttempts": 最大尝试次数（包括初始调用和后续重试），这里是4次
		//"initialBackoff": 初始退避时间，首次重试前等待0.01秒
		//"maxBackoff": 最大退避时间，重试间隔不会超过0.1秒
		//"backoffMultiplier": 退避乘数，每次重试间隔按此倍数递增（这里是2.0，即每次翻倍）
		//"retryableStatusCodes": 可重试的状态码列表，只有UNAVAILABLE状态才会触发重试
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetByID(ctx, &GetByIDRequest{Id: 123})
		cancel()
		require.NoError(t, err)
		t.Log(resp.User)
	}
}

func (s *BalancerTestSuite) TestClientCustomWRR() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`
{
    "loadBalancingConfig": [
        {
            "custom_weighted_round_robin": {}
        }
    ]
}
`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetByID(ctx, &GetByIDRequest{Id: 123})
		cancel()
		require.NoError(t, err)
		t.Log(resp.User)
	}
}

func (s *BalancerTestSuite) TestClientWRR() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`
{
    "loadBalancingConfig": [
        {
            "weighted_round_robin": {}
        }
    ]
}
`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetByID(ctx, &GetByIDRequest{Id: 123})
		cancel()
		require.NoError(t, err)
		t.Log(resp.User)
	}
}

func (s *BalancerTestSuite) TestClient() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`
{
    "loadBalancingConfig": [
        {
            "round_robin": {}
        }
    ]
}
`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetByID(ctx, &GetByIDRequest{Id: 123})
		cancel()
		require.NoError(t, err)
		t.Log(resp.User)
	}
}

func (s *BalancerTestSuite) TestServer() {
	go func() {
		s.startServer(":8090", 10, &Server{
			Name: ":8090",
		})
	}()
	go func() {
		s.startServer(":8091", 20, &Server{
			Name: ":8091",
		})
	}()
	go func() {
		s.startServer(":8092", 30, &Server{
			Name: ":8092",
		})
	}()
	s.startServer(":8093", 30, &FailedServer{
		Name: ":8093",
	})
}

func (s *BalancerTestSuite) startServer(addr string, weight int, svc UserServiceServer) {
	t := s.T()
	em, err := endpoints.NewManager(s.cli, "service/user")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr = "127.0.0.1" + addr
	key := "service/user/" + addr
	l, err := net.Listen("tcp", addr)
	require.NoError(s.T(), err)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 租期
	var ttl int64 = 5
	leaseResp, err := s.cli.Grant(ctx, ttl)
	require.NoError(t, err)

	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息，客户端怎么连你
		Addr: addr,
		Metadata: map[string]any{
			"weight": weight,
		},
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(t, err)
	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		_, err1 := s.cli.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(t, err1)
		//for kaResp := range ch {
		//t.Log(kaResp.String())
		//}
	}()

	server := grpc.NewServer()
	RegisterUserServiceServer(server, svc)
	server.Serve(l)
	kaCancel()
	err = em.DeleteEndpoint(ctx, key)
	if err != nil {
		t.Log(err)
	}
	server.GracefulStop()
	//s.cli.Close()
}

func TestBalancer(t *testing.T) {
	suite.Run(t, new(BalancerTestSuite))
}
