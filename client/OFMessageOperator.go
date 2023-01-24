package main

import (
	"encoding/binary"
	"time"
)

var _ OFMessageI = &OFMessage{}

var OFMessageAnalyser = OFMessage{}

type OFMessage struct {

}

// OFMessageParser to use switch to map messages
func (ofm *OFMessage)OFMessageParser(buf []byte) string{
	str := ""
	messagelength := binary.BigEndian.Uint16(buf[2:4])
	//log.Printf("RecognizingOFMess messagelength:%v buf[1]:%v\n", messagelength, buf[1])
	//header parser
	timeUnix := time.Now().UnixNano()
	switch buf[1] {
	case 10:
		str = str + OFMapper[10](buf,messagelength,timeUnix)
	case 13:
		str = str + OFMapper[13](buf,messagelength,timeUnix)
	case 14:
		str = str + OFMapper[14](buf,messagelength,timeUnix)
	case 18:
		str = str + OFMapper[18](buf,messagelength,timeUnix)
	case 19:
		str = str + OFMapper[19](buf,messagelength,timeUnix)
	default:
		return str
	}
	return str
}

// OFMessageDeparser this function is idle while there is not any specific features.
func (ofm *OFMessage) OFMessageDeparser(buf []byte)  {
	return
}