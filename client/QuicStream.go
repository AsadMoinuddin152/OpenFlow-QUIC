package main

import (
	"github.com/lucas-clemente/quic-go"
)

var _ QuicStreamI = &QuicStream{}

type QuicStream struct {
	stream       quic.Stream
	id           int64
	groupId      uint8

}

func NewQuicStream(stream quic.Stream, id int64) QuicStreamI {
	return &QuicStream{stream: stream,
		id:           id,
		groupId:      0,
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
}

func (qs *QuicStream) Read(buf []byte) (int, error) {
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
