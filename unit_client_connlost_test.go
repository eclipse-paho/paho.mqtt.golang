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
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
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

// MQTT 3.1.1 permits 0x80 for a refused subscription; MQTT 3.1 does not.
func Test_client_subackFailureCodeUsesNegotiatedProtocol(t *testing.T) {
	for _, fallback := range []bool{true, false} {
		t.Run(fmt.Sprintf("fallback=%t", fallback), func(t *testing.T) {
			const timeout = 2 * time.Second
			lost := make(chan error, 1)
			brokerDone := make(chan error, 2)
			opts := NewClientOptions().AddBroker("tcp://unused:1883").
				SetAutoReconnect(false).SetKeepAlive(0).SetConnectTimeout(timeout).
				SetConnectionLostHandler(func(_ Client, err error) { lost <- err })
			if !fallback {
				opts.SetProtocolVersion(4)
			}
			attempt := 0
			opts.SetCustomOpenConnectionFn(func(_ *url.URL, _ ClientOptions) (net.Conn, error) {
				clientConn, brokerConn := net.Pipe()
				t.Cleanup(func() { clientConn.Close(); brokerConn.Close() })
				reject := fallback && attempt == 0
				wantVersion := byte(4)
				if fallback && attempt > 0 {
					wantVersion = 3
				}
				attempt++
				go func() {
					brokerDone <- func() error {
						defer brokerConn.Close()
						if err := brokerConn.SetDeadline(time.Now().Add(timeout)); err != nil {
							return err
						}
						packet, err := packets.ReadPacket(brokerConn)
						if err != nil {
							return err
						}
						connect, ok := packet.(*packets.ConnectPacket)
						if !ok || connect.ProtocolVersion != wantVersion {
							return fmt.Errorf("expected CONNECT version %d, got %v", wantVersion, packet)
						}
						connack := packets.NewControlPacket(packets.Connack).(*packets.ConnackPacket)
						if reject {
							connack.ReturnCode = packets.ErrRefusedBadProtocolVersion
						}
						if err := connack.Write(brokerConn); err != nil || reject {
							return err
						}
						packet, err = packets.ReadPacket(brokerConn)
						if err != nil {
							return err
						}
						subscribe, ok := packet.(*packets.SubscribePacket)
						if !ok {
							return fmt.Errorf("expected SUBSCRIBE, got %v", packet)
						}
						suback := packets.NewControlPacket(packets.Suback).(*packets.SubackPacket)
						suback.MessageID = subscribe.MessageID
						suback.ReturnCodes = []byte{0x80}
						if err := suback.Write(brokerConn); err != nil {
							return err
						}
						packet, err = packets.ReadPacket(brokerConn)
						if fallback {
							if !errors.Is(err, io.EOF) {
								return fmt.Errorf("expected client to close malformed connection, got packet %v, error %v", packet, err)
							}
							return nil
						}
						if _, ok := packet.(*packets.DisconnectPacket); !ok || err != nil {
							return fmt.Errorf("expected DISCONNECT after valid SUBACK, got packet %v, error %v", packet, err)
						}
						return nil
					}()
				}()
				return clientConn, nil
			})
			c := NewClient(opts).(*client)
			defer c.Disconnect(100)
			if token := c.Connect(); !token.WaitTimeout(timeout) || token.Error() != nil {
				t.Fatalf("connect failed: %v", token.Error())
			}
			token := c.Subscribe("topic/a", 0, nil).(*SubscribeToken)
			if !token.WaitTimeout(timeout) {
				t.Fatal("SUBSCRIBE token did not complete")
			}
			if fallback {
				if !errors.Is(token.Error(), ErrMalformedSuback) {
					t.Fatalf("expected ErrMalformedSuback after fallback to MQTT 3.1, got %v", token.Error())
				}
				if len(token.Result()) != 0 {
					t.Fatalf("malformed SUBACK populated results: %v", token.Result())
				}
				select {
				case err := <-lost:
					if !errors.Is(err, ErrMalformedSuback) {
						t.Fatalf("connection lost with unexpected error: %v", err)
					}
				case <-time.After(timeout):
					t.Fatal("malformed SUBACK did not close the connection")
				}
				if _, ok := c.getToken(token.messageID).(*DummyToken); !ok {
					t.Fatal("malformed SUBACK did not release its message ID")
				}
			} else {
				if token.Error() != nil || token.Result()["topic/a"] != 0x80 {
					t.Fatalf("expected valid subscription refusal, got results %v, error %v", token.Result(), token.Error())
				}
				if !c.IsConnected() {
					t.Fatal("valid SUBACK closed the connection")
				}
				c.Disconnect(1000)
			}
			for i := 0; i < attempt; i++ {
				select {
				case err := <-brokerDone:
					if err != nil {
						t.Fatal(err)
					}
				case <-time.After(timeout):
					t.Fatal("broker did not finish")
				}
			}
		})
	}
}
