package main

import (
	"context"
	"log"
	"sync"

	"github.com/lucas-clemente/quic-go"
)
// QuicServerConnection mixed a lots of useless code due to the earlier EPM under
// designing. It should be deleted when the EPM design was ripe.
var _ QuicClientConnectionI = &QuicClientConnection{}

type QuicClientConnection struct {
	Session quic.Session

	ClientStream      map[int64]QuicStreamI
	ClientStreamMutex sync.RWMutex

	Add        string
	State      bool
	Controller ConnectionControllerI
	Uds        UnixClientI
}

func NewQuicClientConnection(session quic.Session, path string) QuicClientConnectionI {
	c := &QuicClientConnection{
		Session:           session,
		Add:               session.RemoteAddr().String(),
		State:             true,
		ClientStreamMutex: sync.RWMutex{},
		ClientStream:      make(map[int64]QuicStreamI),
		Uds:               NewUnixClient(path),
	}
	c.Controller = NewConnectionController(c, &c.ClientStream, &c.ClientStreamMutex, session.ReturnSessionRttStats())
	return c
}

// ConnectionRun run this connection main goroutine
func (qcc *QuicClientConnection) ConnectionRun() {
	qcc.Uds.CreateUnixConnection()
	go qcc.Controller.RunController(qcc.Uds)
}

// ActiveOpenStream OpenStream if needed
// In the pass quic-go version, I don't know why QUIC stream will block the
// program if it has no data to send.
// so, the side of Opening Stream should write data first and the opposite should read
// first. Maybe it is fixed already. Then here can be deleted.
// 	"_, _ = stream.Write(buf)"
//	"_, _ = stream.Read(buf)"
func (qcc *QuicClientConnection) ActiveOpenStream() (QuicStreamI, error) {
	if qcc.State == false {
		return nil, ClientHasBeenClose
	}
	stream, err := qcc.Session.OpenStreamSync(context.Background())
	if err != nil {
		log.Printf("server stream open panic %v\n", err)
		return nil, ClientHasBeenClose
	}
	s := qcc.StreamProcessor(stream)
	return s, err
}

// PassiveOpenStream this should be always listening the new QUIC streams that will
// be Opened.
func (qcc *QuicClientConnection) PassiveOpenStream() (QuicStreamI, error) {
	if qcc.State == false {
		return nil, ClientHasBeenClose
	}
	stream, err := qcc.Session.AcceptStream(context.Background())
	if err != nil {
		log.Printf("server stream accept panic\n")
		return nil, err
	}
	s := qcc.StreamProcessor(stream)
	return s, err
}

func (qcc *QuicClientConnection) StreamProcessor(stream quic.Stream) QuicStreamI {
	streamId := int64(stream.StreamID().StreamNum())
	qcc.ClientStreamMutex.Lock()
	buf := make([]byte, 1)
	_, _ = stream.Write(buf)
	_, _ = stream.Read(buf)
	qs := NewQuicStream(stream, streamId)
	qcc.ClientStream[streamId] = qs
	qcc.ClientStreamMutex.Unlock()
	return qs
}

// CloseClient close this connection, including the ConnectionController
func (qcc *QuicClientConnection) CloseClient() {
	if qcc.State == true {
		qcc.State = false
		qcc.Controller.CloseController()
		log.Printf("CloseClient get")
		i := 0
		for _, v := range qcc.ClientStream {
			_ = v.Close()
			i++
		}
		log.Printf("Client: %s Close, %d Streams close! \n", qcc.Add, i)
		_ = qcc.Session.CloseWithError(0x0, "close by upper application")
	}
}
// ReturnAddr is for logging to add the connection information
func (qcc *QuicClientConnection) ReturnAddr() string {
	return qcc.Session.RemoteAddr().String()
}
