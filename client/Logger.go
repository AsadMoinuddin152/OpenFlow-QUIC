package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

/* here, logs the demanded information when testing */
/* it should be deleted */

var target_src_ip = []byte{10, 0, 0, 1}

var target_dst_ip = []byte{10, 0, 0, 2}

type Logger struct {
	buffer string
	mutex  sync.Mutex
	state  bool
}

func NewLogger() *Logger {
	return &Logger{
		buffer: "",
		mutex:  sync.Mutex{},
		state:  true,
	}
}

// LogRoutine every 20 seconds to write the log to file
func (l *Logger) LogRoutine() {
	for true {
		time.Sleep(20 * time.Second)
		l.mutex.Lock()
		if l.buffer == "" {
			l.mutex.Unlock()
			continue
		}
		log.Printf(l.buffer)
		l.buffer = ""
		if l.state == false {
			l.mutex.Unlock()
			return
		}
		l.mutex.Unlock()
	}
}

func (l *Logger) Close() {
	l.state = false
}

func (l *Logger) Write(format string, a ...interface{}) {
	l.mutex.Lock()
	l.buffer = l.buffer + fmt.Sprintf(format, a...)
	l.mutex.Unlock()
}

// RecognizingOFMess analysis OpenFlow messages with different data bits
func (l *Logger) RecognizingOFMess(buf []byte) {
	l.Write("%s",OFMessageAnalyser.OFMessageParser(buf))
}

