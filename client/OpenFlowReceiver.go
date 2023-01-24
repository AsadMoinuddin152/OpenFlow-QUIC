package main

import (
	"encoding/binary"
	"io"
	"log"
)

// OpenFlowReceiver receives every messages from Uds or quic stream

// WriterRecvFromUds here, accept OpenFlow data from Uds, it should respond to
// read a fully OpenFlow message.
func WriterRecvFromUds(Uds io.Reader, buf []byte) int {
	ret := OFMessageRecvFromUds(Uds, buf[0:8])
	if ret <= 0 {
		log.Printf("WriterRecvFromUds: get8bytes error\n")
		return -1
	}
	SendTotallength := binary.BigEndian.Uint16(buf[2:4])
	if SendTotallength > 8 {
		ret = OFMessageRecvFromUds(Uds, buf[8:SendTotallength])
		if ret != int(SendTotallength-8) {
			log.Printf("WriterRecvFromUds: get got error, out! ret:%d,SendTotallength:%d\n", ret, SendTotallength)
			return -1
		}
	}
	return int(SendTotallength)
}

// OFMessageRecvFromUds get OF data from OVS
func OFMessageRecvFromUds(Uds io.Reader, buf []byte) int {
	length := 0
	bufLength := len(buf)
	for length < bufLength {
		SendDataLength, err := Uds.Read(buf[length:])
		length = SendDataLength + length
		if length == bufLength { //
			return length
		}
		if err != nil {
			log.Printf("RunWriteScheduling: Got Error in Uds by 8 bytes %v \n", err)
			return -1
		}
	}
	return length
}

// OFMessageRecvFromStream here, accept OpenFlow data from a QUIC stream, it should respond to
// read a fully OpenFlow message.
func OFMessageRecvFromStream(s QuicStreamI, buf []byte) (int, error) {
	leng, err := OFMessageRecvAll(s, buf[0:8], 8)
	if err != nil {
		return leng, StreamRecv8Error
	}
	messageLength := int(binary.BigEndian.Uint16(buf[2:4]))
	if messageLength == 8 {
		return leng, nil
	}
	totalLeng, err := OFMessageRecvAll(s, buf[8:messageLength], messageLength-8)
	totalLeng = totalLeng + leng
	if totalLeng != messageLength {
		return -1, StreamRecvOver8Error
	}
	if totalLeng == messageLength && (err != nil || err == StreamRecvError) {
		return totalLeng, err
	}
	return totalLeng, nil
}

// OFMessageRecvAll get OF data from QUIC
func OFMessageRecvAll(s QuicStreamI, buf []byte, length int) (int, error) {
	tempLen := 0
	for tempLen < length {
		rl, err := s.Read(buf[tempLen:])
		tempLen = tempLen + rl
		if err != nil {
			log.Printf("OFMessageRecvAll: stream id %v ReRead OF header failed! %v \n", s.ReturnStreamItSelf().id, err)
			return -1, StreamRecvError
		}
	}
	return tempLen, nil
}
