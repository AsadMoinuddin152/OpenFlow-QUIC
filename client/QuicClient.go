package main

import (
	"crypto/tls"
	"log"
	"sync"

	"github.com/lucas-clemente/quic-go"
)

// QuicClient creates the QuicServer instances, it can map a lots of QuicServer

var _ QuicClientI = &QuicClient{}

type QuicClient struct {
	Server      map[int]QuicServerConnectionI
	ServerMutex sync.Mutex //for multi-controller connection
}

func NewQuicClient() QuicClientI {
	return &QuicClient{
		Server: make(map[int]QuicServerConnectionI),
	}
}

// ConnectToServer add the QuicServerConnection to mapping, it should use mutex
// to avoid the crash. But in the experiment, it only has one connection to controller.
func (q *QuicClient) ConnectToServer(addr string, config *tls.Config, qconfig *quic.Config, fd int) {
	_, ok := q.Server[fd]
	if ok != false {
		panic("QuicClient got another one in mapping with error!!!")
	}
	go func() {
		q.ServerMutex.Lock()
		log.Printf("QCM ConnectToServer pass\n")
		q.Server[fd] = NewQuicServerConnection(addr, fd)
		q.Server[fd].Connect(addr, config, qconfig)
		log.Printf("connected!\n")
		q.ServerMutex.Unlock()
	}()
}

// CloseClient close all Server Connection, but it needs to change it into
// using "fd" to close specify connection
func (q *QuicClient) CloseClient() {
	if len(q.Server) == 0 {
		return
	} else {
		q.ServerMutex.Lock()
		for k, s := range q.Server {
			if s.CheckConnectionState() == 1 {
				s.StopConnecting()
				delete(q.Server, k)
				s = nil
			} else {
				s.CloseControllerByOVS()
				s = nil
			}
		}
		q.ServerMutex.Unlock()
	}
}
