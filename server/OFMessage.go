package main

import (
	"encoding/binary"
	"fmt"
)

// OFMapper logging the corresponding OpenFlow message if the "type" is needed.
// Be careful, these are examples, if the packet is not the type they filters,
// it would cause "SIGSEGV" and "panic". This is unsafe.
var OFMapper = make(map[byte]func([]byte,uint16,int64)string)

func LoadOFMapper() {
	OFMapper[10] = PacketInLogger
	OFMapper[13] = PacketOutLogger
	OFMapper[14] = FlowModLogger
	OFMapper[18] = MultipartRequestLogger
	OFMapper[19] = MultipartReplyLogger
}

func PacketInLogger(buf []byte, leng uint16, nano int64) string {
	Totallength := binary.BigEndian.Uint16(buf[12:14])
	// Ether header -> [messagelength-Totallength]
	EtherHeader := leng - Totallength
	// + src add 26,27,28,29 + dst add 30,31,32,33
	if buf[EtherHeader+12] == 0x08 && buf[EtherHeader+13] == 0x00 && CheckIPAdd(buf[EtherHeader+26:EtherHeader+30], buf[EtherHeader+30:EtherHeader+34]) && (buf[EtherHeader+23] == 0x06 || buf[EtherHeader+23] == 0x11) { // IpHeaderProtocol = EtherHeader +23 TCP == 6 UDP == 17 0x11
		return fmt.Sprintf("OF Message %s size %v IP Address src: %v dst: %v Protocol: %v unixtime: %v \n", "Packet-In", leng, target_src_ip, target_dst_ip, buf[EtherHeader+23], nano)
	}
	return fmt.Sprintf("OF Message %s size %v Protocol: %v unixtime: %v \n", "Packet-In", leng, buf[EtherHeader+23], nano)
}

func PacketOutLogger(buf []byte, leng uint16, nano int64) string {
	//log.Printf("RecognizingOFMess packet-out data:%v\n", buf)
	ActionLength := binary.BigEndian.Uint16(buf[16:18])
	EtherHeader := int(ActionLength + 6 + 18)
	//log.Printf("Log packet-out etherheader:%v ActionLength:%v totallength:%v\n", EtherHeader, ActionLength, messagelength)
	if buf[EtherHeader+12] == 0x08 && buf[EtherHeader+13] == 0x00 && CheckIPAdd(buf[EtherHeader+26:EtherHeader+30], buf[EtherHeader+30:EtherHeader+34]) && (buf[EtherHeader+23] == 0x06 || buf[EtherHeader+23] == 0x11) { // IpHeaderProtocol = EtherHeader +23 TCP == 6 UDP == 17 0x11
		return fmt.Sprintf("OF Message %s size %v IP Address src: %v dst: %v Protocol: %v unixtime: %v \n", "Packet-Out", leng, target_src_ip, target_dst_ip, buf[EtherHeader+23], nano)
	}
	return fmt.Sprintf("OF Message %s size %v Protocol: %v unixtime: %v \n", "Packet-Out", leng, buf[EtherHeader+23], nano)
}

func FlowModLogger(buf []byte, leng uint16, nano int64) string {
	return fmt.Sprintf("OF Message %s size %v unixtime: %v \n", "Flow-mod", leng, nano)
}

func MultipartRequestLogger(buf []byte, leng uint16, nano int64) string {
	return fmt.Sprintf("OF Message %s size %v unixtime: %v \n", "multipart-request", leng, nano)
}

func MultipartReplyLogger(buf []byte, leng uint16, nano int64) string {
	 return fmt.Sprintf("OF Message %s size %v unixtime: %v \n", "multipart-reply", leng, nano)
}