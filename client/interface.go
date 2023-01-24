package main

import "C"
import (
	"crypto/tls"
	"io"

	"github.com/lucas-clemente/quic-go"
)

// UnixServerI start the unix domain socket server to make a connection with OVS
type UnixServerI interface {
	CreateUnixServer()

	CloseConnection()

	io.Reader

	io.Writer
}

// ConnectionControllerI Controls every component, like EPM-algorithm and strategies
type ConnectionControllerI interface {
	// RunWriter control all QUIC streams for running strategies and algorithm to select the target QUIC stream.
	RunWriter()
	// RunReceiver run in all QUIC streams to receive every OpenFlow message and pass to Uds.
	RunReceiver(qs QuicStreamI)
	// CloseController close all components if closing.
	CloseController()
	// AcceptNewStream is running forever to check the new QUIC stream is going to open or not
	AcceptNewStream(n int)
	// RunController initiate the ConnectionController and this goroutine run forever
	RunController(Uds UnixServerI)
}

// QuicClientI used to connect the Quic server, and save QuicServerConnectionI with mapping. */
/* it can support the multicast Quic when it implemented*/
type QuicClientI interface {
	// ConnectToServer initiate everything and connect to controller
	ConnectToServer(addr string, config *tls.Config, qconfig *quic.Config, fd int)
	// CloseClient close all connection, even though the connection is trying to connect.
	CloseClient()
}

// QuicServerConnectionI provide the connection to server object
/* it can support the multicast Quic when it implemented*/
type QuicServerConnectionI interface {
	// Connect to the server (the controller)
	Connect(addr string, config *tls.Config, qconfig *quic.Config)
	// ActiveOpenStream if the switch needs, it can open a new QUIC stream actively.
	ActiveOpenStream() (QuicStreamI, error)
	// PassiveOpenStream
	// Passive Open Stream for server to listen a new stream
	// comes from client executing OpenStreamSync()
	PassiveOpenStream() (QuicStreamI, error)
	// CloseControllerByOVS close this server connection and delete object
	CloseControllerByOVS()

	CheckConnectionState() int

	StopConnecting()
}

// QuicStreamI provide a simple package to QUIC stream, if possible, the new texture (groupID) will be installed to
// use by algorithm on scheduling.
type QuicStreamI interface {
	ChangeStreamGroupId(groupId uint8)

	// ReturnStreamItSelf provide a pointer to access the real QUIC stream's content, like retransmissionQueue, deadline
	ReturnStreamItSelf() *QuicStream

	Read(buf []byte) (int, error)

	Write(data []byte) (int, error)

	Close() error
	// RetMaxFrameLen return the maxframelen to algorithm, only use once in a new OpenFlow message
	RetMaxFrameLen() int
	// RetQueueLen return this stream's queuing length
	RetQueueLen() int
}

// OFMessageI is for coding and decoding the OpenFlow message into logging
type OFMessageI interface {
	OFMessageParser(buf []byte) string

	OFMessageDeparser(buf []byte)
}

// OFHandlerI provide the strategies for sending and receiving
/* do what ever you want to do in here, the strategies are storing in here. */
/* the arguments and returns should be change if it needs*/
type OFHandlerI interface {
	OFMessageRecvOperation(message []byte, Uds UnixServerI) (int, error)

	OFMessageSendOperation(message []byte)
}
