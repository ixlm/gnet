package main

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/logging"

	// "github.com/panjf2000/gnet/v2"
	byteBufPool "github.com/panjf2000/gnet/v2/pkg/pool/bytebuffer"
	goPool "github.com/panjf2000/gnet/v2/pkg/pool/goroutine"
)

// 更多关于gnet的例子参见 gnet-examples
type echoServer struct {
	gnet.BuiltinEventEngine
	engine     gnet.Engine
	addr       string
	multicore  bool
	async      bool
	network    string
	connected  int32
	workerPool *goPool.Pool
}

//---interface EventHandler begin---------------------------------

// OnBoot fires when the engine is ready for accepting connections.
// The parameter engine has information and various utilities.
func (s *echoServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	logging.Infof("running server on %s with multicore=%t", fmt.Sprintf("%s://%s", s.network, s.addr), s.multicore)
	s.engine = eng
	return
}

// OnOpen fires when a new connection has been opened.
//
// The Conn c has information about the connection such as its local and remote addresses.
// The parameter out is the return value which is going to be sent back to the peer.
// Sending large amounts of data back to the peer in OnOpen is usually not recommended.
func (s *echoServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	//设置自定义编解码函数

	//设置连接状态
	atomic.AddInt32(&s.connected, 1)
	return

}

// OnClose fires when a connection has been closed.
// The parameter err is the last known connection error.
func (s *echoServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	if err != nil {
		logging.Errorf("error occurred on close, %v\n", err)
	}
	if s.network != "udp" {
		//c.Context() should == c
	}
	// if disconnected := atomic.StoreInt32(&s.connected, 0); disconnected
	if s.connected == 1 {
		atomic.StoreInt32(&s.connected, 0)
		action = gnet.Shutdown
		s.workerPool.Release()
	}
	return
}

// OnTraffic fires when a socket receives data from the peer.
//
// Note that the []byte returned from Conn.Peek(int)/Conn.Next(int) is not allowed to be passed to a new goroutine,
// as this []byte will be reused within event-loop after OnTraffic() returns.
// If you have to use this []byte in a new goroutine, you should either make a copy of it or call Conn.Read([]byte)
// to read data into your own []byte, then pass the new []byte to the new goroutine.
func (s *echoServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	if s.async { //异步
		//Peek 并不会移动游标
		tempBuf, _ := c.Peek(-1)
		fmt.Printf("%s\n", string(tempBuf))

		buf := byteBufPool.Get()
		len, err := c.WriteTo(buf) //会自动把游标向前推移, Peek不会

		if err != nil {
			return
		}
		logging.Infof("received msg(size:=%d): %s\n", len, buf.String())
		// if buf.String() == "connect" {
		if strings.HasPrefix(buf.String(), "connect") {
			newC, err := s.engine.Dial("tcp", "localhost:16001")
			if err == nil {
				_ = s.workerPool.Submit(func() {
					_ = newC.AsyncWrite([]byte("hello, i'm a client connection"), func(c gnet.Conn, err error) error {
						if c.RemoteAddr() != nil {
							logging.Infof("con=%s, send successful: %v", c.RemoteAddr().String(), err)
						} else {
							logging.Errorf("send failed! target is not online")
						}

						return nil
					})
				})
			}
		}
		//使用完了要放回去，否则会内存泄露
		byteBufPool.Put(buf)
		//这里异步提交，避免阻塞主线程
		_ = s.workerPool.Submit(func() {
			_ = c.AsyncWrite([]byte("thank you"), func(c gnet.Conn, err error) error {
				if c.RemoteAddr() != nil {
					logging.Infof("con=%s, send successful: %v", c.RemoteAddr().String(), err)
				} else {
					logging.Errorf("send failed! target is not online")
				}
				return nil
			})
		})

	}
	// c.Discard(c.InboundBuffered())
	// c.Next(-1)
	return
}

// OnShutdown fires when the engine is being shut down, it is called right after
// all event-loops and connections are closed.
func (s *echoServer) OnShutdown(eng gnet.Engine) {
	logging.Infof("shutdown ....")
}

// OnTick fires immediately after the engine starts and will fire again
// following the duration specified by the delay return value.
func (s *echoServer) OnTick() (delay time.Duration, action gnet.Action) {
	return
}

//---interface EventHandler end---------------------------------

func main() {
	port := 16000
	ss := &echoServer{
		network:    "tcp",
		addr:       fmt.Sprintf(":%d", port),
		multicore:  true,
		async:      true,
		workerPool: goPool.Default(),
	}
	err := gnet.Run(ss,
		ss.network+"://"+ss.addr,
		gnet.WithMulticore(ss.multicore),
		gnet.WithLockOSThread(ss.async),
		gnet.WithTicker(true),
		gnet.WithTCPKeepAlive(1*time.Minute),
		gnet.WithTCPNoDelay(1),
		gnet.WithReuseAddr(true),
		gnet.WithReusePort(true))
	logging.Infof("server exits with error: %v", err)
}
