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
	"time"

	"github.com/ory/graceful"
)

const (
	Worker = "worker"
	Master = "master"
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

	if IsWorker() {
		return c.runAsWorker()
	}
	return c.runAsMaster()
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
	var errch = make(chan error, 1)
	var fn = func() {
		command = exec.Command(os.Args[0])
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.ExtraFiles = append(command.ExtraFiles, c.fdFile)
		command.Env = append(command.Env, fmt.Sprintf("mode=%s", Worker))

		_ = command.Start()
		if err := command.Wait(); err != nil {
			errch <- err
		}
	}
	go fn()

	// keep alived
	var ticker = time.NewTicker(time.Second)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			if command.ProcessState != nil && command.ProcessState.Exited() {
				go fn()
			}
		}
	}()

	var sigch = make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)

	select {
	case <-errch:
		return nil
	case sig := <-sigch:
		_ = command.Process.Signal(os.Interrupt)

		switch sig {
		case os.Interrupt, syscall.SIGTERM:
			return c.listener.Close()
		case syscall.SIGUSR1:
			go fn()
		}
	}
	return nil
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

// IsMaster 当前进程的环境是否是master进程
func IsMaster() bool {
	return !IsWorker()
}

// IsWorkerMode 当前进程的环境是否是worker进程
func IsWorker() bool {
	return os.Getenv("mode") == string(Worker)
}
