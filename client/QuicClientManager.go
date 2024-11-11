package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"github.com/quic-go/quic-go"
)

type QuicClientManager struct {
	Client  QuicClientI
	logSet  bool
	initSet bool
	fd      int
}

func NewQuicClientManager() *QuicClientManager {
	return &QuicClientManager{
		Client:  nil,
		logSet:  false,        // test the logger is set or not
		initSet: false,        // test the QUIC connection is new or already existing
		fd:      GenNumber(3), // random number
	}
}
// InitClient initiates structures and logging
func (q *QuicClientManager) InitClient() {
	if q.initSet == false {
		q.Client = NewQuicClient()
		if q.logSet == false {

			logFile, logErr := os.OpenFile(logaddr, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
			if logErr != nil {
				fmt.Println("Fail to log", *logFile, ", Client start Failed")
				os.Exit(1)
			}
			log.SetOutput(logFile)
			log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)
			q.logSet = true
			/* here, logs the demanded information when testing */
			/* it should be deleted */
			LOG = NewLogger()
			go LOG.LogRoutine()
		}
		q.initSet = true

	}
}

func (q *QuicClientManager) Connect(addr string, tlsConf *tls.Config, conf *quic.Config) int {
	fd := q.Retfd()
	q.Client.ConnectToServer(addr, tlsConf, conf, fd)
	return fd // return immediately if it is still connecting

}

// Retfd random number plus one to prevent from reusing
// the abandoned unix domain socket
func (q *QuicClientManager) Retfd() int {
	q.fd = q.fd + 1
	return q.fd
}

func (q *QuicClientManager) CloseCleanClient() {
	q.Client.CloseClient()
	q.logSet = false
	q.initSet = false
	// LOG.Close()
	// LOG = nil
}
