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

type _alive struct {
	targets map[string]map[string]map[string]*Target
}

var (
	logger   = logging.MustGetLogger("golobo")
	group    sync.Mutex
	alive    = &_alive{make(map[string]map[string]map[string]*Target, 100)} // service >> version >> [target]
	once     sync.Once
	initinal = false
)

const (
	path = "/golobo/rpc/server/lists"
)

type Event struct {
	target *Target
}

type MyLog struct{}

func (l *MyLog) Printf(msg string, args ...interface{}) {
	//	logger.Info(msg, args)
}

func Init(servers []string, ttl int, _error chan error) {

	conn, _, err := initPath(servers)
	if err != nil {
		logger.Error(err)
		_error <- err
		return
	}
	defer conn.Close()

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
						if _, ok := alive.targets[target.Service]; !ok {
							alive.targets[target.Service] = make(map[string]map[string]*Target, 100)
						}
						if _, ok := alive.targets[target.Service][target.Version]; !ok {
							alive.targets[target.Service][target.Version] = make(map[string]*Target, 100)
						}
						if _, ok := alive.targets[target.Service][target.Version][target.String()]; !ok {
							go tunnel.ConnecToServer(fmt.Sprint(target.GetIp(), ":", strconv.FormatInt(target.GetTunnel(), 10)), ttl, &Event{target})
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

func initPath(servers []string) (*zk.Conn, []zk.ACL, error) {

	conn, _, err := zk.Connect(servers, time.Second*2)
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}
	conn.SetLogger(new(MyLog))

	digestACL := zk.DigestACL(zk.PermAll, "shimazaki", "haruka")
	conn.AddAuth(digestACL[0].Scheme, []byte("shimazaki:haruka"))

	if initinal {
		return conn, digestACL, nil
	}
	conn.Create("/golobo", []byte("paruru"), 0, digestACL)
	conn.Create("/golobo/rpc", []byte("paruru"), 0, digestACL)
	conn.Create("/golobo/rpc/server", []byte("paruru"), 0, digestACL)
	conn.Create("/golobo/rpc/server/lists", []byte("paruru"), 0, digestACL)

	once.Do(func() {
		initinal = true
	})
	return conn, digestACL, nil
}

func GetConn(service, version string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithTimeout(time.Second*2))

	addr := alive.get(service, version)
	if addr == "" {
		return nil, errors.New("no server alive now")
	}

	return grpc.Dial(addr, opts...)
}

func (t *_alive) addTarget(target *Target) {
	group.Lock()
	defer group.Unlock()
	t.targets[target.Service][target.Version][target.String()] = target
}

func (t *_alive) removeTarget(target *Target) {
	group.Lock()
	defer group.Unlock()
	delete(t.targets[target.Service][target.Version], target.String())
}

func (t *_alive) get(service, version string) string {
	group.Lock()
	defer group.Unlock()

	if _, ok := t.targets[service]; !ok {
		return ""
	}
	if _, ok := t.targets[service][version]; !ok {
		return ""
	}
	size := len(t.targets[service][version])
	if size == 0 {
		return ""
	}
	size = rand.Intn(size)
	counter := 0
	var addr string

	for _, target := range t.targets[service][version] {
		if counter == size {
			addr = fmt.Sprint(target.GetIp(), ":", strconv.FormatInt(target.GetRpc(), 10))
			break
		}
		counter += 1
	}
	return addr
}

func (e *Event) OnOpen(handle tunnel.Handle) {
	//	logger.Info("addTarget", e.target.String())
	alive.addTarget(e.target)
}

func (e *Event) OnClose(localAddr string, err error) {
	//	logger.Info("removeTarget", e.target.String())
	alive.removeTarget(e.target)
}

func (e *Event) OnMessage(handle tunnel.Handle, payload []byte) {
	// do nothing
}

func _init() {

	go func() {
		for _ = range time.NewTicker(time.Second * 10).C {
			for service, versionMap := range alive.targets {
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

func Publish(servers []string, target *Target) error {

	conn, digestACL, err := initPath(servers)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer conn.Close()

	_, err = conn.Create(fmt.Sprint(path, "/", target.String()), []byte("paruru"), 0, digestACL)
	if err != nil {
		return err
	}
	return nil
}
