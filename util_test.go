package golobo

import (
	"fmt"
	"os"
	"testing"

	"github.com/op/go-logging"
)

func init() {

	format := logging.MustStringFormatter(`[%{module}] %{time:2006-01-02 15:04:05} [%{level}] [%{longpkg} %{shortfile}] { %{message} }`)

	backendConsole := logging.NewLogBackend(os.Stderr, "", 0)
	backendConsole2Formatter := logging.NewBackendFormatter(backendConsole, format)

	logging.SetBackend(backendConsole2Formatter)
	logging.SetLevel(logging.CRITICAL, "tunnel")
}

func TestShow(t *testing.T) {

	servers := []string{"192.168.25.5:2190", "192.168.25.5:2191", "192.168.25.5:2192"}
	targets, err := Show(servers)
	if err != nil {
		logger.Fatal(err)
	}

	for index, value := range targets {
		logger.Info(fmt.Sprint("[", index, "]"), fmt.Sprint("`", value, "`"))
	}
}

func TestRemove(t *testing.T) {

	servers := []string{"192.168.25.5:2190", "192.168.25.5:2191", "192.168.25.5:2192"}
	if err := Remove(servers, `ip:"127.0.0.1" rpc:7772 tunnel:8882 service:"demo" version:"v0.2" `); err != nil {
		logger.Fatal(err)
	}
}
