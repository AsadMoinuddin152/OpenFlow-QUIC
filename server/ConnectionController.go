package main

import (
	"io"
	"log"

	"github.com/quic-go/quic-go/"

	"sync"
	"time"
)

// ConnectionController takes charge of the whole QUIC connection and messages processing.
type ConnectionController struct {
	Client          QuicClientConnectionI
	Streams         *map[int64]QuicStreamI
	StreamsMutex    *sync.RWMutex
	Rtt             *utils.RTTStats
	controllerstate bool
	Uds             UnixClientI
	alg             *Algorithm //algorithm
	ofh             OFHandlerI
	addr            string
}

func NewConnectionController(Client QuicClientConnectionI, ClientStream *map[int64]QuicStreamI, ClientStreamMutex *sync.RWMutex, rtt *utils.RTTStats) ConnectionControllerI {
	return &ConnectionController{Client: Client,
		Streams:         ClientStream,
		StreamsMutex:    ClientStreamMutex,
		alg:             NewAlgorithm(),
		controllerstate: true,
		Rtt:             rtt,
		ofh:             NewOFHandler(),
	}
}

func (c *ConnectionController) CloseController() {
	if c.controllerstate == true {
		c.controllerstate = false
		c.StreamsMutex.Lock()
		for _, s := range *c.Streams {
			_ = s.Close()
		}
		c.Rtt = nil
		c.StreamsMutex.Unlock()
		c.alg.CloseAlg()
		c.Uds.CloseConnection()
		c.Client.CloseClient()
		go func() {
			time.Sleep(5 * time.Second)
			c.Streams = nil
			c.StreamsMutex = nil
			c.Uds = nil
			c.ofh = nil
		}()
	}
}

// RunController main goroutine for running scheduling.
func (c *ConnectionController) RunController(Uds UnixClientI) {
	if c.controllerstate == true {
		c.addr = c.Client.ReturnAddr()
		c.Uds = Uds
		go c.AcceptNewStream()  // hold receiving all quic stream
		c.ActiveCreateStream(5) // open x quic streams
		LoadOFMapper()          //load OpenFlow Message parser function
		c.RunWriter()
	}
}

func (c *ConnectionController) ActiveCreateStream(n int) {
	for i := 0; i < n; i++ {
		if c.controllerstate == false {
			return
		}
		s, err := c.Client.ActiveOpenStream()
		if err != nil {
			c.CloseController()
			log.Printf("ActiveCreateStream failed. stream ID:%d err:%v \n", s.ReturnStreamItSelf().stream.StreamID().StreamNum(), err)
			return
		}
		go c.RunReceiver(s)
	}
}

// CreateNewStream always receives the new coming QUIC stream and add them into main goroutine
func (c *ConnectionController) AcceptNewStream() {
	// hold receiving all quic stream
	for {
		if c.controllerstate == false {
			return
		}
		s, err := c.Client.PassiveOpenStream()
		if err != nil {
			c.CloseController()
			log.Printf("PassiveOpenStream failed. err: %v \n", err)
			return
		}
		go c.RunReceiver(s)
	}
	//
}

// RunWriter accepting every OpenFlow packet from Uds -> RYU
func (c *ConnectionController) RunWriter() {
	// Writer
	SendTotallength := 0
	buf := make([]byte, 65536)
	streamId := int64(0)
	for {
		if c.controllerstate == false {
			return
		}
		SendTotallength = WriterRecvFromUds(c.Uds, buf)
		if SendTotallength <= 0 {
			log.Printf("RunWriteScheduling got error OFMessageRecvFromUds, out!\n")
			c.CloseController()
			return
		}
		// run strategies
		c.ofh.OFMessageSendOperation(buf[0:SendTotallength])
		c.StreamsMutex.RLock()
		// run algorithm
		streamId = c.alg.RunAlgorithm(SendTotallength, c.Streams)
		//go LOG.RecognizingOFMess(c.addr, buf[0:SendTotallength])
		_, err := (*c.Streams)[streamId].Write(buf[0:SendTotallength])
		c.StreamsMutex.RUnlock()
		if err != nil {
			log.Printf("RunWriteScheduling fail while sending something, killing. %v\n", err)
			c.CloseController()
			return
		}
	}
}

// RunReceiver receiving every OpenFlow data from QUIC streams.
// each QUIC stream get one receiver(RunReceiver).
func (c *ConnectionController) RunReceiver(s QuicStreamI) {
	buf := make([]byte, 65536)
	streamID := s.ReturnStreamItSelf().id
	for {
		if c.controllerstate == false {
			return
		}
		leng, err := OFMessageRecvFromStream(s, buf)
		if err != nil { // if occurs error
			if leng == 0 && err == io.EOF {
				continue
			}
			log.Printf("RunReceiver: Stream Read fail! %d bytes, Stream id : %v \n", leng, streamID)
			c.CloseController()
			return
		}
		// run strategies
		uleng, erru := c.ofh.OFMessageRecvOperation(buf[0:leng], c.Uds)
		//go LOG.RecognizingOFMess(c.addr, buf[0:leng])
		if err != nil || uleng == 0 {
			log.Printf("RunRecvScheduling: Stream id : %v got Uds Write error: %v\n", streamID, erru)
			c.CloseController()
			return
		}
	}
}
