package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/amamina/golobo"
	pb "github.com/amamina/golobo/test"
	"github.com/amamina/tunnel"
	"github.com/op/go-logging"
	"github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	logger = logging.MustGetLogger("main")
)

func init() {
	format := logging.MustStringFormatter(`[%{module}] %{time:2006-01-02 15:04:05} [%{level}] [%{longpkg} %{shortfile}] { %{message} }`)

	backendConsole := logging.NewLogBackend(os.Stderr, "", 0)
	backendConsole2Formatter := logging.NewBackendFormatter(backendConsole, format)

	logging.SetBackend(backendConsole2Formatter)
	logging.SetLevel(logging.INFO, "")
	logging.SetLevel(logging.ERROR, "golobo")
}

const (
	path = "/golobo/rpc/server/lists"
)

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
		Rpc:     7778,
		Tunnel:  8002,
		Service: "dummy",
		Version: "v0.2",
	}

	tunnel.Init(5, 2)
	logger.Info(">>>>> ", target.String())
	go func() {
		if err := tunnel.ListenOn(int(target.Tunnel), new(Event)); err != nil {
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
	conn, _, err := zk.Connect(servers, time.Second*2)
	if err != nil {
		logger.Fatal(err)
	}
	defer conn.Close()
	conn.SetLogger(new(MyLog))

	digestACL := zk.DigestACL(zk.PermAll, "shimazaki", "haruka")
	conn.AddAuth(digestACL[0].Scheme, []byte("shimazaki:haruka"))

	conn.Create("/golobo", []byte("paruru's node"), 0, digestACL)
	conn.Create("/golobo/rpc", []byte("paruru's node"), 0, digestACL)
	conn.Create("/golobo/rpc/server", []byte("paruru's node"), 0, digestACL)
	conn.Create("/golobo/rpc/server/lists", []byte("paruru's node"), 0, digestACL)

	msg, err := conn.Create(fmt.Sprint(path, "/", target.String()), []byte("paruru's node"), 0, digestACL)
	if err != nil {
		logger.Error(err)
	}
	logger.Info("create >>", msg)

	wait := make(chan int)
	<-wait
}

func (e *Event) OnOpen(handle tunnel.Handle) {
}

func (e *Event) OnClose(remoteAddr string) {
}

func (e *Event) OnMessage(handle tunnel.Handle, payload []byte) {
}
