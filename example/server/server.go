package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/amamina/golobo"
	pb "github.com/amamina/golobo/example"
	"github.com/amamina/tunnel"
	"github.com/op/go-logging"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	logger = logging.MustGetLogger("server")
)

func init() {
	format := logging.MustStringFormatter(`[%{module}] %{time:2006-01-02 15:04:05} [%{level}] [%{longpkg} %{shortfile}] { %{message} }`)

	backendConsole := logging.NewLogBackend(os.Stderr, "", 0)
	backendConsole2Formatter := logging.NewBackendFormatter(backendConsole, format)

	logging.SetBackend(backendConsole2Formatter)
	logging.SetLevel(logging.INFO, "")
	logging.SetLevel(logging.ERROR, "golobo")
	logging.SetLevel(logging.CRITICAL, "tunnel")
}

type Event struct{}

type MyLog struct{}

func (l *MyLog) Printf(msg string, args ...interface{}) {
	//	logger.Info(msg, args)
}

type DummyServer struct{}

func (s *DummyServer) Hello(ctx context.Context, in *pb.Payload) (*pb.Payload, error) {

	logger.Infof("Receive: %s", in.String())
	return &pb.Payload{Message: "received", Serialkey: 7743, Timestamp: time.Now().Format("2006-01-02 15:04:05")}, nil
}

func main() {

	target := &golobo.Target{
		Ip:      "127.0.0.1",
		Rpc:     7773,
		Tunnel:  8883,
		Service: "demo",
		Version: "v0.1",
	}

	tunnel.InitHeartBeat(15, 3)
	logger.Info(">>>>> ", target.String())
	go func() {
		if err := tunnel.ListenOn(int(target.Tunnel), 3, new(Event)); err != nil {
			logger.Fatal(err)
		}
	}()

	go func() {
		listener, err := net.Listen("tcp", fmt.Sprint(":", strconv.FormatInt(target.Rpc, 10)))
		if err != nil {
			logger.Fatalf("failed to listen: %v", err)
		}

		rpcServer := grpc.NewServer()
		pb.RegisterDummyServer(rpcServer, &DummyServer{})
		// Register reflection service on gRPC server.
		reflection.Register(rpcServer)
		if err := rpcServer.Serve(listener); err != nil {
			logger.Fatalf("failed to serve: %v", err)
		}
	}()

	servers := []string{"192.168.25.5:2190", "192.168.25.5:2191", "192.168.25.5:2192"}
	if err := golobo.Publish(servers, target); err != nil {
		logger.Error(err)
	}

	wait := make(chan int)
	<-wait
}

func (e *Event) OnOpen(handle tunnel.Handle) {
}

func (e *Event) OnClose(remoteAddr string, err error) {
}

func (e *Event) OnMessage(handle tunnel.Handle, payload []byte) {
}
