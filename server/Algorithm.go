package main

import (
	"math"
	"sync"
)

// Algorithm is the scheduler of EPM, which provides the OpenFlow message mapping into QUIC streams
type AlgorithmT struct {
	ModValue      int64
	DivValue      int64
	state         bool // div
	framesCounter int64
	queuingLen    int64
}

type Algorithm struct {
	AlgorithmT   map[int64]*AlgorithmT
	MaxFrameSize int64
	minFrames    *Min
	min          *Min
	algorMutex   sync.RWMutex
}

type Min struct {
	streamId int64
	value    int64
}

func NewAlgorithm() *Algorithm {
	return &Algorithm{
		MaxFrameSize: 1,
		min:          &Min{streamId: 0, value: 1000000},
		minFrames:    &Min{streamId: 0, value: 1000000},
		AlgorithmT:   make(map[int64]*AlgorithmT),
		algorMutex:   sync.RWMutex{},
	}
}

func NewAlgorithmTemp() *AlgorithmT {
	return &AlgorithmT{
		ModValue:      1,
		DivValue:      1,
		state:         false,
		framesCounter: 0,
		queuingLen:    0,
	}
}

func (a *Algorithm) CloseAlg() {
	a.algorMutex.Lock()
	a.AlgorithmT = nil
	a.min = nil
	a.minFrames = nil
	a.algorMutex.Unlock()
}

// InitAlgorithm Update the length of AlgorithmT, it needs the same length as streams
func (a *Algorithm) InitAlgorithm(streams *map[int64]QuicStreamI) {
	a.algorMutex.Lock()
	for k, _ := range *streams {
		a.AlgorithmT[k] = NewAlgorithmTemp()
	}
	a.algorMutex.Unlock()
}

// AlgorithmGetAllQueuingLen Update all Queuing Length of streams
func (a *Algorithm) AlgorithmGetAllQueuingLen(streams *map[int64]QuicStreamI) {
	stream := int64(0)
	maxsize := int64(0)
	if len(*streams) != len(a.AlgorithmT) {
		a.InitAlgorithm(streams)
	}
	a.algorMutex.RLock()
	for k, v := range *streams {
		a.AlgorithmT[k].queuingLen = int64(v.RetQueueLen())
		maxsize = int64(v.RetMaxFrameLen())
		if stream < maxsize {
			stream = maxsize
		}
	}
	a.algorMutex.RUnlock()
	a.MaxFrameSize = maxsize
}

// RunAlgorithm it will return the selected minimum probability of Hol blocking stream
func (a *Algorithm) RunAlgorithm(MessageLen int, streams *map[int64]QuicStreamI) int64 {
	remainSum := int64(0)
	add := int64(0)
	a.AlgorithmGetAllQueuingLen(streams) // check is it the same length
	a.algorMutex.RLock()
	for k, v := range a.AlgorithmT { // search the minimum queuing length stream
		if v.queuingLen == 0 {
			a.algorMutex.RUnlock() // if got one "zero" queuing, return.
			return k
		}
		if a.min.value > v.queuingLen {
			a.min.value = v.queuingLen
			a.min.streamId = k
		}
		v.ModValue = v.queuingLen % a.MaxFrameSize
	}
	for _, v := range a.AlgorithmT { // find the same round as the minimum queuing streams
		v.state = math.Ceil(float64(v.queuingLen)/float64(a.MaxFrameSize)) == math.Ceil(float64(a.min.value)/float64(a.MaxFrameSize))
		// find the streams which contains the rest data length larger than 128(minFrameSize)
		if (v.ModValue < a.MaxFrameSize-128) && v.state == true {
			remainSum += a.MaxFrameSize - v.ModValue
			add++
		}
	}
	for k, v := range a.AlgorithmT {
		if v.state == true { // calculate which one gets the minimum frames
			v.framesCounter = (v.queuingLen+int64(MessageLen)-(remainSum-a.MaxFrameSize+v.ModValue))/a.MaxFrameSize + add - 1
			if a.minFrames.value > v.framesCounter {
				a.minFrames.value = v.framesCounter
				a.minFrames.streamId = k
			}
		}
	}
	a.algorMutex.RUnlock()
	return a.minFrames.streamId
}
