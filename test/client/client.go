package main

import (
	"context"
	"os"
	"runtime/debug"
	"time"

	"github.com/amamina/golobo"
	pb "github.com/amamina/golobo/test"
	"github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("test")
)

func init() {

	format := logging.MustStringFormatter(`[%{module}] %{time:2006-01-02 15:04:05} [%{level}] [%{longpkg} %{shortfile}] { %{message} }`)

	backendConsole := logging.NewLogBackend(os.Stderr, "", 0)
	backendConsole2Formatter := logging.NewBackendFormatter(backendConsole, format)

	logging.SetBackend(backendConsole2Formatter)
	logging.SetLevel(logging.CRITICAL, "tunnel")
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			logger.Error(string(debug.Stack()))
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	_error := make(chan error)

	servers := []string{"192.168.25.5:2190", "192.168.25.5:2191", "192.168.25.5:2192"}
	path := "/golobo/rpc/server/lists"

	go golobo.Init(servers, "digest", []byte("shimazaki:haruka"), path, _error)
	go func() {
		select {
		case <-ctx.Done():
			return
		case err := <-_error:
			logger.Error(err)
			cancel()
		}
	}()

	go func() {
		for _ = range time.NewTicker(time.Second * 10).C {
			conn, err := golobo.GetConn("dummy", "v0.2")
			if err != nil {
				logger.Fatal(err)
			}
			defer conn.Close()
			dummyClient := pb.NewDummyClient(conn)

			payload, err := dummyClient.Hello(context.Background(), &pb.Payload{Message: "dummy", Timestamp: time.Now().Format("2006-01-02 15:04:05")})
			if err != nil {
				logger.Fatal(err)
			}

			logger.Infof("Response: %s", payload.String())
		}
	}()

	<-ctx.Done()
	logger.Error(ctx.Err())
}
