/*
 * Copyright (c) 2021 IBM Corp and others.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * and Eclipse Distribution License v1.0 which accompany this distribution.
 *
 * The Eclipse Public License is available at
 *    https://www.eclipse.org/legal/epl-2.0/
 * and the Eclipse Distribution License is available at
 *   http://www.eclipse.org/org/documents/edl-v10.php.
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

package mqtt

import (
	"bytes"
	"errors"
	"log/slog"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// syncBuffer is a goroutine-safe buffer. internalConnLost performs its work (and
// logging) across several goroutines, so the test needs synchronised access to
// the captured log output.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// Test_internalConnLost_NoBugLogWhenAutoReconnectDisabled is a regression test for
// https://github.com/eclipse-paho/paho.mqtt.golang/issues/697.
//
// When AutoReconnect is disabled, ConnectionLost is called with willReconnect=false
// and so the connection-lost handler legitimately returns a nil reconnect function.
// internalConnLost previously logged "BUG BUG BUG reconnection function is nil" in
// this case, even though it is the expected behaviour. This test ensures that log is
// not emitted when no reconnect was ever expected.
func Test_internalConnLost_NoBugLogWhenAutoReconnectDisabled(t *testing.T) {
	logBuf := &syncBuffer{}
	logger := slog.New(slog.NewTextHandler(logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	opts := NewClientOptions().
		SetAutoReconnect(false).
		SetCleanSession(true).
		SetConnectionLostHandler(nil). // keep the test isolated from the default handler
		SetLogger(logger)
	c := NewClient(opts).(*client)

	// Place the client in a state that mimics an active connection so that
	// internalConnLost runs to completion rather than exiting early (it returns
	// immediately when c.conn is nil).
	netClient, netServer := net.Pipe()
	defer netClient.Close()
	defer netServer.Close()
	c.conn = netClient
	c.stop = make(chan struct{})
	c.commsStopped = make(chan struct{})
	close(c.commsStopped) // comms are already stopped so stopCommsWorkers can finish immediately
	c.status.forceConnectionStatus(connected)

	c.internalConnLost(errors.New("simulated connection loss"))

	// internalConnLost completes asynchronously. "internalConnLost complete" is the
	// final log line, emitted after the spurious-bug check, so it is a reliable
	// signal that the whole flow has run.
	deadline := time.Now().Add(2 * time.Second)
	for !strings.Contains(logBuf.String(), "internalConnLost complete") {
		if time.Now().After(deadline) {
			t.Fatalf("internalConnLost did not complete in time; log so far:\n%s", logBuf.String())
		}
		time.Sleep(time.Millisecond)
	}

	if got := logBuf.String(); strings.Contains(got, "BUG BUG BUG") {
		t.Fatalf("spurious bug log emitted when AutoReconnect is disabled:\n%s", got)
	}
}
