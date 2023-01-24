package main

import "C"
import (
	"context"
	"log"
	"sync"

	"github.com/lucas-clemente/quic-go"
)

// QuicServer creates the QuicServer instances, it can map a lots of QuicClient

var _ QuicServerI = &QuicServer{}

type QuicServer struct {
	Listener quic.Listener
	// Client
	Client       map[string]QuicClientConnectionI
	ClientMutex  sync.Mutex
	connectState bool
}

func NewQuicServer() QuicServerI {
	return &QuicServer{
		Listener:     nil,
		Client:       make(map[string]QuicClientConnectionI),
		ClientMutex:  sync.Mutex{},
		connectState: false,
	}
}

func (s *QuicServer) CheckListenerAlive() int {
	if s.Listener != nil {
		return 0
	}
	return 1
}

// ServerRun if a new QUIC connection came, then it creates the new Unix domain
// socket one-by-one. And this is run forever until the listener is down.
func (s *QuicServer) ServerRun(listener quic.Listener, UnixPath string) error {
	s.Listener = listener
	s.Client = make(map[string]QuicClientConnectionI)
	s.ClientMutex = sync.Mutex{}
	for {
		log.Printf("listener working \n")
		session, err := s.Listener.Accept(context.Background())
		s.connectState = true
		if err != nil {
			log.Printf("BlockServer error %v, Listener is unavailable!!!\n", err)
			return QuicListenerUnavailable
		}
		ClientAddr := session.RemoteAddr().String()
		log.Printf("listener get %v \n", ClientAddr)
		s.ClientMutex.Lock()
		client := NewQuicClientConnection(session, UnixPath)
		s.Client[ClientAddr] = client
		client.ConnectionRun()
		s.ClientMutex.Unlock()
	}
}

func (s *QuicServer) CloseListener() {
	err := s.Listener.Close()
	if err != nil {
		return
	}
	s.Listener = nil
}

func (s *QuicServer) CloseAllConnection() {
	s.ClientMutex.Lock()
	if s.connectState == true {
		for k, client := range s.Client {
			client.CloseClient()
			delete(s.Client, k)
		}
	}
	s.ClientMutex.Unlock()
}

func (s *QuicServer) CloseClientConnection(Addr string) error {
	s.ClientMutex.Lock()
	c, err := s.Client[Addr]
	if err != true {
		s.ClientMutex.Unlock()
		return CanNotFindClientByMap
	}
	c.CloseClient()
	delete(s.Client, Addr)
	s.ClientMutex.Unlock()
	return nil
}
