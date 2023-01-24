package main

import (
	"time"
)

// OFHandlerOperatorReceiver is the executer. Every OpenFlow message will match the function if needs.
// Here is the EPM's extra operations to improve the OpenFlow protocol performance.

// OFHandlerMapperRecv
// every functions add into mapping for executing.
var OFHandlerMapperRecv = make(map[byte]func([]byte, *OFHandler, UnixClientI) (int, error))

// LoadOFHandlerMapperRecv initiates the mapping
func LoadOFHandlerMapperRecv() {
	OFHandlerMapperRecv[10] = PacketInOperatorRecv
	OFHandlerMapperRecv[13] = PacketOutOperatorRecv
	OFHandlerMapperRecv[14] = FlowModOperatorRecv
	OFHandlerMapperRecv[18] = MultipartRequestOperatorRecv
	OFHandlerMapperRecv[19] = MultipartReplyOperatorRecv
}

func MultipartReplyOperatorRecv(message []byte, ofh *OFHandler, Uds UnixClientI) (int, error) {
	leng, err := Uds.Write(message)
	return leng, err
}

func MultipartRequestOperatorRecv(message []byte, ofh *OFHandler, Uds UnixClientI) (int, error) {
	if ofh.PacketIn == true {
		time.Sleep(10 * time.Millisecond) //1RTT, 5ms link latency in experiment.
	}
	leng, err := Uds.Write(message)
	return leng, err
}

func FlowModOperatorRecv(message []byte, ofh *OFHandler, Uds UnixClientI) (int, error) {
	leng, err := Uds.Write(message)
	return leng, err
}

func PacketOutOperatorRecv(message []byte, ofh *OFHandler, Uds UnixClientI) (int, error) {
	leng, err := Uds.Write(message)
	return leng, err
}

func PacketInOperatorRecv(message []byte, ofh *OFHandler, Uds UnixClientI) (int, error) {
	leng, err := Uds.Write(message)
	return leng, err
}

func DefaultOperatorRecv(message []byte, ofh *OFHandler, Uds UnixClientI) (int, error) {
	leng, err := Uds.Write(message)
	return leng, err
}
