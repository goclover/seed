package seed

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ory/graceful"
)

type Mode string

const (
	ModeWorker Mode = "worker"
	ModeMaster Mode = "master"
)

type MSeed interface {
	Router

	// HTTPServer 返回实例的 *http.Server
	HTTPServer() *http.Server

	// 设置默认的http server
	SetHTTPServer(srv *http.Server) MSeed

	// Run 启动HTTPServer
	//
	// 	addrs 监听的端口,非必选，但如果没有通过HTTPServer pointer修改且值不给可能导致服务启动失败
	Run(addrs ...string) error

	// Run 启动HTTPServer
	//
	// 	certFile https certFile
	// 	keyFile https keyFile
	// 	addrs 监听的端口,非必选，但如果没有通过HTTPServer pointer修改且值不给可能导致服务启动失败
	RunTLS(certFile, keyFile string, addrs ...string) error

	// Mode 返回当前程序的运行环境（master进程还是worker进程）
	Mode() Mode
}

// mseed is an app driven by Router
type mseed struct {
	Router

	once      sync.Once
	certFile  string
	keyFile   string
	enableTLS bool

	listener net.Listener
	fdFile   *os.File
	server   *http.Server
}

// New return *mseed
func New() MSeed {
	return &mseed{
		Router: NewRouter(),
		server: &http.Server{},
	}
}

func (c *mseed) HTTPServer() *http.Server {
	return c.server
}

func (c *mseed) SetHTTPServer(srv *http.Server) MSeed {
	c.server = srv
	return c
}

// start http server
func (c *mseed) Run(addrs ...string) error {
	if len(addrs) > 0 {
		c.server.Addr = addrs[0]
	}
	if c.server.Handler == nil {
		c.server.Handler = c
	}

	var f = c.runAsMaster
	if os.Getenv("mode") != "" {
		f = c.runAsWorker
	}
	return f()
}

func (c *mseed) RunTLS(certFile, keyFile string, addrs ...string) error {
	c.certFile = certFile
	c.keyFile = keyFile
	c.enableTLS = true
	return c.Run(addrs...)
}

func (c *mseed) runAsMaster() (err error) {
	// master thread listen tcp port
	c.once.Do(func() {
		if c.listener, err = net.Listen("tcp", c.server.Addr); err != nil {
			return
		}
		var ok bool
		var t *net.TCPListener
		if t, ok = c.listener.(*net.TCPListener); !ok {
			err = errors.New("tcp protocal support only")
			return
		}
		if c.fdFile, err = t.File(); err != nil {
			return
		}
	})
	if err != nil {
		return
	}

	var command *exec.Cmd
	var fn = func() {
		command = exec.Command(os.Args[0])
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.ExtraFiles = append(command.ExtraFiles, c.fdFile)
		command.Env = append(command.Env, fmt.Sprintf("mode=%s", ModeWorker))
		command.Start()
	}
	go fn()

	var ch = make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)

	for sig := range ch {
		_ = command.Process.Signal(os.Interrupt)
		_ = command.Process.Release()

		switch sig {
		case os.Interrupt, syscall.SIGTERM:
			return c.listener.Close()
		case syscall.SIGUSR1:
			go fn()
		}
	}
	return nil
}

func (c *mseed) Mode() Mode {
	if os.Getenv("mode") == string(ModeWorker) {
		return ModeWorker
	}
	return ModeMaster
}

func (c *mseed) runAsWorker() (err error) {
	var f = os.NewFile(uintptr(3), "connection")
	var l net.Listener

	if l, err = net.FileListener(f); err != nil {
		return err
	}
	var fn = func() error {
		if c.enableTLS {
			return c.server.ServeTLS(l, c.certFile, c.keyFile)
		}
		return c.server.Serve(l)
	}
	return graceful.Graceful(fn, c.server.Shutdown)
}
