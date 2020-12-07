/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package util

import (
	"fmt"
	"io"
	"log"
	"strings"

	"devt.de/krotik/common/datautil"
)

// Loger with loglevel support
// ===========================

/*
LogLevel represents a logging level
*/
type LogLevel string

/*
Log levels
*/
const (
	Debug LogLevel = "debug"
	Info           = "info"
	Error          = "error"
)

/*
LogLevelLogger is a wrapper around loggers to add log level functionality.
*/
type LogLevelLogger struct {
	logger Logger
	level  LogLevel
}

/*
NewLogLevelLogger wraps a given logger and adds level based filtering functionality.
*/
func NewLogLevelLogger(logger Logger, level string) (*LogLevelLogger, error) {
	llevel := LogLevel(strings.ToLower(level))

	if llevel != Debug && llevel != Info && llevel != Error {
		return nil, fmt.Errorf("Invalid log level: %v", llevel)
	}

	return &LogLevelLogger{
		logger,
		llevel,
	}, nil
}

/*
Level returns the current log level.
*/
func (ll *LogLevelLogger) Level() LogLevel {
	return ll.level
}

/*
LogError adds a new error log message.
*/
func (ll *LogLevelLogger) LogError(m ...interface{}) {
	ll.logger.LogError(m...)
}

/*
LogInfo adds a new info log message.
*/
func (ll *LogLevelLogger) LogInfo(m ...interface{}) {
	if ll.level == Info || ll.level == Debug {
		ll.logger.LogInfo(m...)
	}
}

/*
LogDebug adds a new debug log message.
*/
func (ll *LogLevelLogger) LogDebug(m ...interface{}) {
	if ll.level == Debug {
		ll.logger.LogDebug(m...)
	}
}

// Logging implementations
// =======================

/*
MemoryLogger collects log messages in a RingBuffer in memory.
*/
type MemoryLogger struct {
	*datautil.RingBuffer
}

/*
NewMemoryLogger returns a new memory logger instance.
*/
func NewMemoryLogger(size int) *MemoryLogger {
	return &MemoryLogger{datautil.NewRingBuffer(size)}
}

/*
LogError adds a new error log message.
*/
func (ml *MemoryLogger) LogError(m ...interface{}) {
	ml.RingBuffer.Add(fmt.Sprintf("error: %v", fmt.Sprint(m...)))
}

/*
LogInfo adds a new info log message.
*/
func (ml *MemoryLogger) LogInfo(m ...interface{}) {
	ml.RingBuffer.Add(fmt.Sprintf("%v", fmt.Sprint(m...)))
}

/*
LogDebug adds a new debug log message.
*/
func (ml *MemoryLogger) LogDebug(m ...interface{}) {
	ml.RingBuffer.Add(fmt.Sprintf("debug: %v", fmt.Sprint(m...)))
}

/*
Slice returns the contents of the current log as a slice.
*/
func (ml *MemoryLogger) Slice() []string {
	sl := ml.RingBuffer.Slice()
	ret := make([]string, len(sl))
	for i, lm := range sl {
		ret[i] = lm.(string)
	}
	return ret
}

/*
Reset resets the current log.
*/
func (ml *MemoryLogger) Reset() {
	ml.RingBuffer.Reset()
}

/*
Size returns the current log size.
*/
func (ml *MemoryLogger) Size() int {
	return ml.RingBuffer.Size()
}

/*
String returns the current log as a string.
*/
func (ml *MemoryLogger) String() string {
	return ml.RingBuffer.String()
}

/*
StdOutLogger writes log messages to stdout.
*/
type StdOutLogger struct {
	stdlog func(v ...interface{})
}

/*
NewStdOutLogger returns a stdout logger instance.
*/
func NewStdOutLogger() *StdOutLogger {
	return &StdOutLogger{log.Print}
}

/*
LogError adds a new error log message.
*/
func (sl *StdOutLogger) LogError(m ...interface{}) {
	sl.stdlog(fmt.Sprintf("error: %v", fmt.Sprint(m...)))
}

/*
LogInfo adds a new info log message.
*/
func (sl *StdOutLogger) LogInfo(m ...interface{}) {
	sl.stdlog(fmt.Sprintf("%v", fmt.Sprint(m...)))
}

/*
LogDebug adds a new debug log message.
*/
func (sl *StdOutLogger) LogDebug(m ...interface{}) {
	sl.stdlog(fmt.Sprintf("debug: %v", fmt.Sprint(m...)))
}

/*
NullLogger discards log messages.
*/
type NullLogger struct {
}

/*
NewNullLogger returns a null logger instance.
*/
func NewNullLogger() *NullLogger {
	return &NullLogger{}
}

/*
LogError adds a new error log message.
*/
func (nl *NullLogger) LogError(m ...interface{}) {
}

/*
LogInfo adds a new info log message.
*/
func (nl *NullLogger) LogInfo(m ...interface{}) {
}

/*
LogDebug adds a new debug log message.
*/
func (nl *NullLogger) LogDebug(m ...interface{}) {
}

/*
BufferLogger logs into a buffer.
*/
type BufferLogger struct {
	buf io.Writer
}

/*
NewBufferLogger returns a buffer logger instance.
*/
func NewBufferLogger(buf io.Writer) *BufferLogger {
	return &BufferLogger{buf}
}

/*
LogError adds a new error log message.
*/
func (bl *BufferLogger) LogError(m ...interface{}) {
	fmt.Fprintln(bl.buf, fmt.Sprintf("error: %v", fmt.Sprint(m...)))
}

/*
LogInfo adds a new info log message.
*/
func (bl *BufferLogger) LogInfo(m ...interface{}) {
	fmt.Fprintln(bl.buf, fmt.Sprintf("%v", fmt.Sprint(m...)))
}

/*
LogDebug adds a new debug log message.
*/
func (bl *BufferLogger) LogDebug(m ...interface{}) {
	fmt.Fprintln(bl.buf, fmt.Sprintf("debug: %v", fmt.Sprint(m...)))
}
