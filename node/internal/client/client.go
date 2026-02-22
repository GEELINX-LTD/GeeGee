package client

import (
	"context"
	"log"

	pb "github.com/geelinx-ltd/geegee/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	serverAddr   string
	conn         *grpc.ClientConn
	probeClient  pb.ProbeServiceClient
	streamClient pb.ProbeService_ReportMetricsClient
}

func NewGrpcClient(serverAddr string) *GrpcClient {
	return &GrpcClient{
		serverAddr: serverAddr,
	}
}

func (c *GrpcClient) Connect() error {
	log.Printf("Connecting to controller at %s...", c.serverAddr)

	// 这里使用非安全连接用于测试环境，生产环境中建议切换为 TLS 并进行 Token 鉴定
	conn, err := grpc.NewClient(c.serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	c.conn = conn
	c.probeClient = pb.NewProbeServiceClient(conn)

	// 建立双向数据流长连接
	ctx := context.Background() // 或者使用有生命周期的 context
	stream, err := c.probeClient.ReportMetrics(ctx)
	if err != nil {
		return err
	}
	c.streamClient = stream

	// 启动一个 goroutine 接收主控端的下发指令（如更新探测目标等）
	go c.receiveLoop()

	log.Println("Successfully connected and established metrics stream.")
	return nil
}

func (c *GrpcClient) receiveLoop() {
	for {
		resp, err := c.streamClient.Recv()
		if err != nil {
			log.Printf("Error receiving from stream: %v", err)
			return // 流断开，实际应实现退避重连逻辑
		}

		if !resp.Success {
			log.Printf("Server returned error: %s", resp.Message)
		} else if len(resp.ProbeTargets) > 0 {
			log.Printf("Received new probe targets: %d targets. (Mocking applied)", len(resp.ProbeTargets))
			// TODO: 将下发的 Target 通知到 network/tcpping collector 组件
		}
	}
}

func (c *GrpcClient) SendMetrics(req *pb.ReportRequest) error {
	if c.streamClient == nil {
		return nil
	}
	// 在高频上报时，我们直接向流写入即可，得益于 gRPC 流，TCP 层面复用而且基于 Protobuf，非常高效
	err := c.streamClient.Send(req)
	if err != nil {
		log.Printf("Failed to push metrics to stream: %v", err)
		return err
	}
	log.Printf("Successfully pushed aggregated packet for Node %s", req.NodeId)
	return nil
}

func (c *GrpcClient) Close() {
	if c.streamClient != nil {
		c.streamClient.CloseSend()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
