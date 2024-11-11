package main

import "C"
import (
	"github.com/quic-go/quic-go"
	"io"
)

// UnixClientI start the unix domain socket server to make a connection with OVS
type UnixClientI interface {
	CreateUnixConnection()

	CloseConnection()

	io.Reader

	io.Writer
}

// ConnectionControllerI Controls every component, like EPM-algorithm and strategies
type ConnectionControllerI interface {
	// ActiveCreateStream launch the QUIC streams first
	ActiveCreateStream(n int)
	// AcceptNewStream is running forever to check the new QUIC stream is going to open or not
	AcceptNewStream()
	// RunWriter control all QUIC streams for running strategies and algorithm to select the target QUIC stream.
	RunWriter()
	// CloseController close all components if closing.
	CloseController()
	// RunReceiver run in all QUIC streams to receive every OpenFlow message and pass to Uds.
	RunReceiver(qs QuicStreamI)
	// RunController initiate the ConnectionController and this goroutine run forever
	RunController(Uds UnixClientI)
}
// QuicServerI used to connect the Quic server, and save QuicClientConnectionI with mapping. */
/* it can support the multicast Quic when it implemented*/
type QuicServerI interface {

	CheckListenerAlive() int

	CloseListener()
	// CloseAllConnection close all addresses
	CloseAllConnection()
	// CloseClientConnection close the specific address
	CloseClientConnection(UnixPath string) error
	// ServerRun initiate QUIC listener and ready for preparing new QUIC connection.
	ServerRun(listener quic.Listener, UnixPath string) error
}

// QuicClientConnectionI provide the connection to client object
/* it can support the multicast Quic when it implemented*/
type QuicClientConnectionI interface {

	CloseClient()
	// ActiveOpenStream
	// For Server open a new bidirectional stream to notice client.
	ActiveOpenStream() (QuicStreamI, error)
	// PassiveOpenStream
	// Passive Open Stream for server to listen a new stream
	// comes from client executing OpenStreamSync()
	PassiveOpenStream() (QuicStreamI, error)
	// ConnectionRun handle quic stream
	ConnectionRun()

	ReturnAddr() string
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

	OFMessageRecvOperation(message []byte, Uds UnixClientI) (int, error)

	OFMessageSendOperation(message []byte)
}
