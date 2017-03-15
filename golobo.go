package golobo

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/amamina/tunnel"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"github.com/samuel/go-zookeeper/zk"
	"google.golang.org/grpc"
)

var (
	logger = logging.MustGetLogger("golobo")
	group  sync.Mutex
	alive  = make(map[string]map[string]map[string]*Target, 100) // service >> version >> [target]
)

type Event struct {
	target *Target
}

type MyLog struct{}

func (l *MyLog) Printf(msg string, args ...interface{}) {
	//	logger.Info(msg, args)
}

func Init(servers []string, scheme string, auth []byte, path string, _error chan error) {

	conn, _, err := zk.Connect(servers, time.Second*2)
	if err != nil {
		logger.Error(err)
		_error <- err
		return
	}
	defer conn.Close()
	conn.SetLogger(new(MyLog))
	conn.AddAuth(scheme, auth)

	var ctx context.Context
	var cancel context.CancelFunc

	for {
		list, _, event, err := conn.ChildrenW(path)
		if err != nil {
			logger.Error(err)
			_error <- err
			return
		}

		targetList := make([]*Target, len(list))
		for index, v := range list {
			target := new(Target)
			if err = proto.UnmarshalText(v, target); err != nil {
				logger.Error(err)
				_error <- err
				return
			}
			targetList[index] = target
		}

		if cancel != nil {
			cancel()
		}

		ctx, cancel = context.WithCancel(context.Background())
		go func() {
			ticker := time.NewTicker(time.Second * 5)
			go func() {
				for _ = range ticker.C {
					for _, target := range targetList {
						if _, ok := alive[target.Service]; !ok {
							alive[target.Service] = make(map[string]map[string]*Target, 100)
						}
						if _, ok := alive[target.Service][target.Version]; !ok {
							alive[target.Service][target.Version] = make(map[string]*Target, 100)
						}
						if _, ok := alive[target.Service][target.Version][target.String()]; !ok {
							go tunnel.ConnecToServer(fmt.Sprint(target.GetIp(), ":", strconv.FormatInt(target.GetTunnel(), 10)), &Event{target})
						}
					}
				}
			}()
			<-ctx.Done()
			ticker.Stop()
		}()

		<-event
	}

}

func GetConn(service, version string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithTimeout(time.Second*2))

	err := errors.New("no server alive now")

	if _, ok := alive[service]; !ok {
		return nil, err
	}

	if _, ok := alive[service][version]; !ok {
		return nil, err
	}

	size := len(alive[service][version])
	if size == 0 {
		return nil, err
	}
	size = rand.Intn(size)
	counter := 0
	var addr string

	for _, target := range alive[service][version] {
		if counter == size {
			addr = fmt.Sprint(target.GetIp(), ":", strconv.FormatInt(target.GetRpc(), 10))
			break
		}
		counter += 1
	}

	if addr == "" {
		return nil, err
	}
	return grpc.Dial(addr, opts...)
}

func addTarget(target *Target) {
	group.Lock()
	defer group.Unlock()
	alive[target.Service][target.Version][target.String()] = target
}

func removeTarget(target *Target) {
	group.Lock()
	defer group.Unlock()
	delete(alive[target.Service][target.Version], target.String())
}

func (e *Event) OnOpen(handle tunnel.Handle) {
	addTarget(e.target)
}

func (e *Event) OnClose(localAddr string) {
	removeTarget(e.target)
}

func (e *Event) OnMessage(handle tunnel.Handle, payload []byte) {
	// do nothing
}

func _init() {

	go func() {
		for _ = range time.NewTicker(time.Second * 10).C {
			for service, versionMap := range alive {
				for version, servers := range versionMap {
					for server, _ := range servers {
						logger.Info(service, version, server)
					}
				}
			}
			logger.Info("------------------------------------------")
		}
	}()
}
