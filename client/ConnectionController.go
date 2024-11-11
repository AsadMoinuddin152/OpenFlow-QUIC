package main

import (
	"io"
	"log"
	"sync"
	"time"

	"github.com/quic-go/quic-go/internal/utils"
)

// ConnectionController takes charge of the whole QUIC connection and messages processing.
var _ ConnectionControllerI = &ConnectionController{}

type ConnectionController struct {
	qsc             QuicServerConnectionI
	Streams         *map[int64]QuicStreamI
	StreamsMutex    *sync.RWMutex
	Rtt             *utils.RTTStats
	controllerstate bool
	Uds             UnixServerI
	alg             *Algorithm //algorithm
	ofh             OFHandlerI
}

func NewConnectionController(qsc QuicServerConnectionI, ss *map[int64]QuicStreamI, sm *sync.RWMutex, rtt *utils.RTTStats) ConnectionControllerI {
	return &ConnectionController{
		qsc:             qsc,
		Streams:         ss,
		StreamsMutex:    sm,
		Rtt:             rtt,
		controllerstate: true,
		alg:             NewAlgorithm(),
		ofh:             NewOFHandler(),
	}
}

func (c *ConnectionController) CloseController() {
	if c.controllerstate == true {
		c.controllerstate = false
		log.Printf("Close Controller\n")
		c.StreamsMutex.Lock()
		for _, s := range *c.Streams {
			_ = s.Close()
		}
		c.Rtt = nil
		c.StreamsMutex.Unlock()
		go func() { // make sure all goroutines quitted then clean the pointer
			time.Sleep(5 * time.Second)
			c.alg.CloseAlg()
			c.alg = nil
			c.Streams = nil
			c.StreamsMutex = nil
		}()
		c.ofh = nil

	}
	log.Printf("Closed Controller\n")
}

// RunController main goroutine for running scheduling.
func (c *ConnectionController) RunController(Uds UnixServerI) {
	if c.controllerstate == true {
		c.Uds = Uds
		c.AcceptNewStream(5)
		go c.AcceptNewStream(0) // hold receiving all quic stream
		LoadOFMapper()          //load OpenFlow Message parser function
		c.RunWriter()
	}
}

// AcceptNewStream always receives the new coming QUIC stream and add them into main goroutine
func (c *ConnectionController) AcceptNewStream(n int) {
	if n == 0 {
		for { // run forever
			if c.controllerstate == true {

				s, err := c.qsc.PassiveOpenStream()
				c.StreamsMutex.Lock()
				c.alg.InitAlgorithm(c.Streams)
				c.StreamsMutex.Unlock()
				if err != nil {
					log.Printf("RunWriteScheduling got error, can not open a new stream in PassiveOpenStream. \n")
					c.CloseController()
					return
				}
				go c.RunReceiver(s)
			}
		}
	} else { //accepting the excepted number of QUIC streams when initiating
		for i := 0; i < n; i++ {
			if c.controllerstate == true {
				s, err := c.qsc.PassiveOpenStream()
				c.StreamsMutex.Lock()
				c.alg.InitAlgorithm(c.Streams)
				c.StreamsMutex.Unlock()
				if err != nil {
					log.Printf("RunWriteScheduling got error, can not open a new stream in PassiveOpenStream. \n")
					c.CloseController()
					return
				}
				go c.RunReceiver(s)
			}
		}
	}
}

// RunWriter accepting every OpenFlow packet from Uds -> OVS
func (c *ConnectionController) RunWriter() {
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
		//go LOG.RecognizingOFMess(buf[0:SendTotallength])
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
		//go LOG.RecognizingOFMess(buf[0:leng])
		if err != nil || uleng == 0 {
			log.Printf("RunRecvScheduling: Stream id : %v got Uds Write error: %v\n", streamID, erru)
			c.CloseController()
			return
		}
	}
}
