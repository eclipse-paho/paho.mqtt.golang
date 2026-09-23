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
	"fmt"
	"io"
	"testing"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// Test_resumePublishIndependentHeader - Issue 703 - Checks that the header is not altered when being resent
// because that could lead to a data race
func Test_resumePublishIndependentHeader(t *testing.T) {
	for _, qos := range []byte{0, 1, 2} {
		t.Run(fmt.Sprintf("qos%d", qos), func(t *testing.T) {
			c := NewClient(NewClientOptions()).(*client)
			c.persist.Open()
			defer c.persist.Close()
			c.obound = make(chan *PacketAndToken, 1)
			p := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
			p.Qos, p.Retain, p.MessageID = qos, true, 42
			p.TopicName, p.Payload = "resume/test", []byte("payload")
			c.persist.Put(outboundKeyFromMID(p.MessageID), p)

			// A second reconnect must also leave the stored packet untouched.
			for range 2 {
				c.resume(false, nil)
				out := <-c.obound
				resumed := out.p.(*packets.PublishPacket)
				if qos != 0 && resumed == p { // QOS 0 messages are unchanged so do not need to be copied
					t.Fatal("resume reused the stored packet")
				}
				var wire bytes.Buffer
				if err := resumed.Write(&wire); err != nil {
					t.Fatal(err)
				}
				decoded, err := packets.ReadPacket(&wire)
				if err != nil {
					t.Fatal(err)
				}
				sent := decoded.(*packets.PublishPacket)
				if sent.Qos != qos || sent.Dup != (qos != 0) || !sent.Retain ||
					sent.TopicName != p.TopicName || !bytes.Equal(sent.Payload, p.Payload) {
					t.Errorf("unexpected resumed packet: %v", sent)
				}
				if qos != 0 && sent.MessageID != p.MessageID {
					t.Errorf("message ID changed: got %d, want %d", sent.MessageID, p.MessageID)
				}
				if out.t.(*PublishToken).MessageID() != p.MessageID {
					t.Error("publish token message ID changed")
				}
				if p.Dup || p.RemainingLength != 0 {
					t.Errorf("stored header was modified: %v", p.FixedHeader)
				}
			}
		})
	}
}

// Test_resumePublishConcurrentWrite (intended for use with --race). Check that we can write a packet
// whilst concurrently resuming the connection.
func Test_resumePublishConcurrentWrite(t *testing.T) {
	c := NewClient(NewClientOptions()).(*client)
	c.persist.Open()
	defer c.persist.Close()
	c.obound = make(chan *PacketAndToken, 1)
	p := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
	p.Qos, p.MessageID = 1, 42
	p.TopicName, p.Payload = "resume/test", bytes.Repeat([]byte("x"), 4096)
	c.persist.Put(outboundKeyFromMID(p.MessageID), p)

	// A publish started after connectionUp can be both in the store and in
	// the outgoing worker while resume is processing the store.
	started := make(chan struct{})
	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		close(started)
		for {
			select {
			case <-stop:
				done <- nil
				return
			default:
				if err := p.Write(io.Discard); err != nil {
					done <- err
					return
				}
			}
		}
	}()
	defer func() {
		close(stop)
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()
	<-started
	for i := 0; i < 1000; i++ {
		c.resume(false, nil)
		out := <-c.obound
		if err := out.p.Write(io.Discard); err != nil {
			t.Fatal(err)
		}
	}
}
