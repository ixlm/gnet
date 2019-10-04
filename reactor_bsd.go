// Copyright 2019 Andy Pan. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build darwin netbsd freebsd openbsd dragonfly

package gnet

import (
	"github.com/panjf2000/gnet/internal"
	"github.com/panjf2000/gnet/netpoll"
)

func (svr *server) activateMainReactor() {
	defer svr.signalShutdown()

	_ = svr.mainLoop.poller.Polling(func(fd int, filter int16, job internal.Job) error {
		if fd == 0 {
			return job()
		}
		return svr.acceptNewConnection(fd)
	})
}

func (svr *server) activateSubReactor(lp *loop) {
	defer svr.signalShutdown()

	if lp.idx == 0 && svr.opts.Ticker {
		go lp.loopTicker()
	}

	_ = lp.poller.Polling(func(fd int, filter int16, job internal.Job) error {
		if fd == 0 {
			return job()
		}

		c := lp.connections[fd]
		if filter == netpoll.EVFilterWrite && !c.outboundBuffer.IsEmpty() {
			return lp.loopOut(c)
		} else if filter == netpoll.EVFilterRead {
			return lp.loopIn(c)
		} else if filter == netpoll.EVFilterSock {
			return lp.loopCloseConn(c, nil)
		}

		return nil
	})
}
