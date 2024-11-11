package main

import (
	"context"
	"crypto/tls"
	"log"

	//"log"
	"strconv"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

// QuicServerConnection mixed a lots of useless code due to the earlier EPM under
// designing. It should be deleted when the EPM design was ripe.
var _ QuicServerConnectionI = &QuicServerConnection{}

type QuicServerConnection struct {
	tryConnect  bool
	connected   bool
	stopConnect bool
	Addr        string
	Session     quic.Session

	Streams      map[int64]QuicStreamI
	StreamsMutex sync.RWMutex
	cc           ConnectionControllerI

	Uds UnixServerI

	fd int
}

func NewQuicServerConnection(Addr string, fd int) QuicServerConnectionI {
	return &QuicServerConnection{
		tryConnect:   false,
		connected:    false,
		Addr:         Addr,
		Session:      nil,
		Streams:      make(map[int64]QuicStreamI),
		StreamsMutex: sync.RWMutex{},
		stopConnect:  false,
		fd:           fd,
	}
}

// Connect Open a new QUIC connection. Also, if the connection is created, the Uds Server
// (Unix domain Socket) will be opened immediately for OVS to connect.
func (qsc *QuicServerConnection) Connect(addr string, config *tls.Config, qconfig *quic.Config) {
	qsc.Uds = NewUnixServer("/tmp/mininet" + strconv.Itoa(qsc.fd) + ".sock")
	log.Printf("New UnixServer pass\n")
	for i := 0; i < 50; i++ { // retry 50 times every 300ms
		session, err := quic.DialAddr(addr, config, qconfig)
		if qsc.stopConnect == true {
			log.Printf("connecting stopped! Addr : %v \n", addr)
			return
		}
		if err == nil { // if connected, here, it will provide the Uds to OVS
			qsc.Session = session
			qsc.Uds.CreateUnixServer()
			log.Printf("CreateUnixServer created!\n")
			qsc.cc = NewConnectionController(qsc, &qsc.Streams, &qsc.StreamsMutex, session.ReturnSessionRttStats())
			qsc.connected = true

			qsc.cc.RunController(qsc.Uds) // Run ConnectionController

			if qsc.stopConnect == true {
				log.Printf("connected! but still needs to stop! \n")
				qsc.CloseControllerByOVS()
				return
			}
			return
		}
		time.Sleep(time.Duration(i*300) * time.Millisecond)
	}

}

// ActiveOpenStream OpenStream if needed
// In the pass quic-go version, I don't know why QUIC stream will block the
// program if it has no data to send.
// so, the side of Opening Stream should write data first and the opposite should read
// first. Maybe it is fixed already. Then here can be deleted.
// 	"_, _ = stream.Write(buf)"
//	"_, _ = stream.Read(buf)"
func (qsc *QuicServerConnection) ActiveOpenStream() (QuicStreamI, error) {
	if qsc.stopConnect == true {
		log.Printf("ActiveOpenStream is opening a broken connection! \n")
		return nil, UsingBrokenConnection
	}
	stream, err := qsc.Session.OpenStreamSync(context.Background())
	if err != nil {
		log.Printf("ActiveOpenStream stream open panic\n")
		return nil, err
	}
	s := qsc.StreamProcessor(stream)
	return s, err
}

// PassiveOpenStream this should be always listening the new QUIC streams that will
// be Opened.
func (qsc *QuicServerConnection) PassiveOpenStream() (QuicStreamI, error) {
	if qsc.stopConnect == true {
		log.Printf("PassiveOpenStream is opening a broken connection! \n")
		return nil, UsingBrokenConnection
	}
	stream, err := qsc.Session.AcceptStream(context.Background())
	if err != nil {
		log.Printf("PassiveOpenStream stream accept fail\n")
		return nil, err
	}
	s := qsc.StreamProcessor(stream)
	return s, err
}

func (qsc *QuicServerConnection) StreamProcessor(stream quic.Stream) QuicStreamI {
	streamId := int64(stream.StreamID().StreamNum())
	qsc.StreamsMutex.Lock()
	buf := make([]byte, 1)
	_, _ = stream.Write(buf)
	_, _ = stream.Read(buf)
	s := NewQuicStream(stream, streamId)
	qsc.Streams[streamId] = s
	qsc.StreamsMutex.Unlock()
	return s
}

// StopConnecting set the stopConnect bit, when the upper application close the connection.
func (qsc *QuicServerConnection) StopConnecting() {
	qsc.stopConnect = true
}

func (qsc *QuicServerConnection) CheckConnectionState() int {
	if qsc.connected == true {
		return 0 // 0 -> connected
	}
	return 1 // 1 -> not connected
}

func (qsc *QuicServerConnection) CloseControllerByOVS() {
	qsc.cc.CloseController()
	// clean the memory
	qsc.Uds.CloseConnection()
	qsc.Uds = nil
	go func() {
		time.Sleep(5 * time.Second)
		qsc.cc = nil
		qsc.Streams = nil
	}()

	qsc.tryConnect = false
	qsc.stopConnect = true
	_ = qsc.Session.CloseWithError(0x0, "Client kill the connection")
}
