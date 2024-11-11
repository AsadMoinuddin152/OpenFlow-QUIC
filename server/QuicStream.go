package main

import (
	"github.com/quic-go/quic-go"
)

type QuicStream struct {
	stream  quic.Stream
	id      int64
	groupId uint8
}

func NewQuicStream(stream quic.Stream, id int64) QuicStreamI {
	return &QuicStream{stream: stream,
		id:      id,
		groupId: 0,
	}
}

func (qs *QuicStream) ChangeStreamGroupId(groupId uint8) {
	qs.groupId = groupId
}

func (qs *QuicStream) ReturnStreamItSelf() *QuicStream {
	return qs
}

func (qs *QuicStream) Write(data []byte) (int, error) {
	return qs.stream.Write(data)
	//log.Printf("Stream.Write: stream id : %v Write data :%v,len: %d, err: %v\n", qs.id, data, l, err)
}

func (qs *QuicStream) Read(buf []byte) (int, error) {
	//qs.RecvPacket++
	return qs.stream.Read(buf)
}

func (qs *QuicStream) RetMaxFrameLen() int {
	return qs.stream.MaxFrameLen()
}

func (qs *QuicStream) RetQueueLen() int {
	return qs.stream.QueuingLen()
}

func (qs *QuicStream) Close() error {
	return qs.stream.Close()
}
