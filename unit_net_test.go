/*
 * Copyright (c) 2026 IBM Corp and others.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * and Eclipse Distribution License v1.0 which accompany this distribution.
 *
 * The Eclipse Public License is available at
 *    https://www.eclipse.org/legal/epl-2.0/
 * and the Eclipse Distribution License is available at
 *   http://www.eclipse.org/org/documents/edl-v10.php.
 */

package mqtt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

func Test_connectMQTT_rejectsOversizedConnack(t *testing.T) {
	// The broker response advertises a two-byte body but does not send it. A
	// one-byte limit must reject the packet before attempting to read the body.
	conn := bytes.NewBuffer([]byte{0x20, 0x02})
	connectPacket := packets.NewControlPacket(packets.Connect).(*packets.ConnectPacket)

	_, _, err := connectMQTT(conn, connectPacket, 4, noopSLogger, 1)
	if !errors.Is(err, packets.ErrPacketTooLarge) {
		t.Fatalf("expected ErrPacketTooLarge, got %v", err)
	}
}

func Test_startIncomingComms_rejectsOversizedPacket(t *testing.T) {
	conn := bytes.NewBuffer([]byte{0x30, 0xff, 0xff, 0xff, 0x7f})
	inboundFromStore := make(chan packets.ControlPacket)
	close(inboundFromStore)

	output := startIncomingComms(
		conn,
		&testCommsFns{maxIncomingPacketSize: 1024},
		inboundFromStore,
		noopSLogger,
	)

	select {
	case result := <-output:
		if !errors.Is(result.err, packets.ErrPacketTooLarge) {
			t.Fatalf("expected ErrPacketTooLarge, got %v", result.err)
		}
	case <-time.After(time.Second):
		t.Fatal("startIncomingComms did not report the oversized packet")
	}
}

// SUBACKs must contain one valid return code per subscription, with failure
// (0x80) supported only by MQTT 3.1.1 (including its bridge mode).
func Test_startIncomingComms_subackReturnCodes(t *testing.T) {
	const messageID = 1

	type testCase struct {
		name          string
		subs          []string
		returnCodes   []byte
		expectedError bool
		dummyToken    bool
	}
	cases := []testCase{
		{
			name:          "qos0",
			subs:          []string{"topic/a"},
			returnCodes:   []byte{0},
			expectedError: false,
		},
		{
			name:          "multiple",
			subs:          []string{"topic/a", "topic/b"},
			returnCodes:   []byte{0, 0},
			expectedError: false,
		},
		{
			name:          "too_many",
			subs:          []string{"topic/a"},
			returnCodes:   []byte{0, 1},
			expectedError: true,
		},
		{
			name:          "too_few",
			subs:          []string{"topic/a", "topic/b"},
			returnCodes:   []byte{0},
			expectedError: true,
		},
		{
			name:          "empty",
			subs:          []string{"topic/a"},
			returnCodes:   []byte{},
			expectedError: true,
		},
		{name: "qos1", subs: []string{"topic/a"}, returnCodes: []byte{1}},
		{name: "qos2", subs: []string{"topic/a"}, returnCodes: []byte{2}},
		{name: "denied", subs: []string{"topic/a"}, returnCodes: []byte{0x80}},
		{name: "mixed_valid", subs: []string{"topic/a", "topic/b", "topic/c", "topic/d"}, returnCodes: []byte{0, 1, 2, 0x80}},
		{name: "invalid_first", subs: []string{"topic/a", "topic/b"}, returnCodes: []byte{0xfe, 1}, expectedError: true},
		{name: "invalid_middle", subs: []string{"topic/a", "topic/b", "topic/c"}, returnCodes: []byte{0, 0xfe, 2}, expectedError: true},
		{name: "invalid_last", subs: []string{"topic/a", "topic/b"}, returnCodes: []byte{0, 0xfe}, expectedError: true},
		{name: "unknown_id_valid", returnCodes: []byte{2}, dummyToken: true},
		{name: "unknown_id_invalid", returnCodes: []byte{0xfe}, expectedError: true, dummyToken: true},
	}
	for _, code := range []byte{3, 4, 0x7f, 0x81, 0xfe, 0xff} {
		cases = append(cases, testCase{
			name: fmt.Sprintf("reserved_%02x", code), subs: []string{"topic/a"},
			returnCodes: []byte{code}, expectedError: true,
		})
	}

	for _, protocolVersion := range []uint{3, 4, 0x83, 0x84} {
		for _, c := range cases {
			t.Run(fmt.Sprintf("protocol_%02x/%s", protocolVersion, c.name), func(t *testing.T) {
				expectedError := c.expectedError
				if protocolVersion == 3 || protocolVersion == 0x83 {
					for _, code := range c.returnCodes {
						if code == 0x80 {
							expectedError = true
						}
					}
				}
				// Set the topics requested in the matching SUBSCRIBE.
				token := newToken(packets.Subscribe).(*SubscribeToken)
				token.subs = c.subs

				// Build the SUBACK to process.
				suback := packets.NewControlPacket(packets.Suback).(*packets.SubackPacket)
				suback.MessageID = messageID
				suback.ReturnCodes = c.returnCodes
				var conn bytes.Buffer
				if err := suback.Write(&conn); err != nil {
					t.Fatalf("failed to write suback: %v", err)
				}

				inboundFromStore := make(chan packets.ControlPacket) // Store unused in this test
				close(inboundFromStore)

				// Start the incoming processor, the SUBACK is already on conn
				comms := &testCommsFns{token: token, protocolVersion: protocolVersion}
				if c.dummyToken {
					comms.token = &DummyToken{}
				}
				output := startIncomingComms(&conn, comms, inboundFromStore, noopSLogger)

				// Regardless of the result, the token should be done
				if !c.dummyToken {
					select {
					case <-token.Done():
					case <-time.After(time.Second):
						t.Fatalf("subscribe token was not completed")
					}
				}

				// capture everything from the channel to ensure it's fully drained
				var received []incomingComms
			drainOutput:
				for {
					select {
					case msg, ok := <-output:
						if !ok {
							break drainOutput
						}
						received = append(received, msg)
					case <-time.After(time.Second):
						t.Fatalf("startIncomingComms did not complete")
					}
				}

				if !reflect.DeepEqual(comms.freedIDs, []uint16{messageID}) {
					t.Errorf("expected message ID to be freed once, got %v", comms.freedIDs)
				}
				if expectedError {
					malformedSubackErrors := 0
					for _, msg := range received {
						if errors.Is(msg.err, ErrMalformedSuback) {
							malformedSubackErrors++
						}
					}
					if malformedSubackErrors != 1 {
						t.Errorf("expected ErrMalformedSuback once on chan (sub: %v, codes: %v), got %d in %v", c.subs, c.returnCodes, malformedSubackErrors, received)
					}
					if !c.dummyToken && !errors.Is(token.Error(), ErrMalformedSuback) {
						t.Errorf("expected ErrMalformedSuback (sub: %v, codes: %v), got %v", c.subs, c.returnCodes, token.Error())
					}
					if len(token.Result()) != 0 {
						t.Errorf("malformed SUBACK populated results: %v", token.Result())
					}
				} else {
					if len(received) != 1 || !errors.Is(received[0].err, io.EOF) {
						t.Errorf("expected normal closure, got %v", received)
					}
					if token.Error() != nil {
						t.Errorf("expected successful SUBACK (sub: %v, codes: %v), got %v", c.subs, c.returnCodes, token.Error())
					}
					if !c.dummyToken {
						want := make(map[string]byte, len(c.subs))
						for i, topic := range c.subs {
							want[topic] = c.returnCodes[i]
						}
						if !reflect.DeepEqual(token.Result(), want) {
							t.Errorf("expected results %v, got %v", want, token.Result())
						}
					}
				}
			})
		}
	}
}

// testCommsFns is a basic implementation of commsFns for use with startIncomingComms
type testCommsFns struct {
	token                 tokenCompletor
	maxIncomingPacketSize uint32
	protocolVersion       uint
	freedIDs              []uint16
}

func (c *testCommsFns) getToken(uint16) tokenCompletor {
	return c.token
}

func (c *testCommsFns) freeID(id uint16) {
	c.freedIDs = append(c.freedIDs, id)
}

func (c *testCommsFns) UpdateLastReceived() {}

func (c *testCommsFns) UpdateLastSent() {}

func (c *testCommsFns) getWriteTimeOut() time.Duration {
	return 0
}

func (c *testCommsFns) getMaxIncomingPacketSize() uint32 {
	return c.maxIncomingPacketSize
}

func (c *testCommsFns) getProtocolVersion() uint {
	return c.protocolVersion
}

func (c *testCommsFns) persistOutbound(packets.ControlPacket) {}

func (c *testCommsFns) persistInbound(packets.ControlPacket) {}

func (c *testCommsFns) pingRespReceived() {}
