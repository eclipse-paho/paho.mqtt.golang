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

package packets

import (
	"bytes"
	"encoding/hex"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// Independent wire fixtures cover every packet type and all PUBLISH QoS levels.
var writeFixtures = []struct {
	name string
	wire string
}{
	{"connect", "100f00044d5154540402003c0003616263"},
	{"connack", "20020100"},
	{"publish_qos0", "30050001616869"},
	{"publish_qos1_dup_retained", "3b07000161002a6869"},
	{"publish_qos2", "3407000161002a6869"},
	{"publish_multibyte_length", "308301000161" + strings.Repeat("78", 128)},
	{"puback", "4002002a"},
	{"pubrec", "5002002a"},
	{"pubrel", "6202002a"},
	{"pubcomp", "7002002a"},
	{"subscribe", "820a002a0001610100016202"},
	{"suback", "9005002a000180"},
	{"unsubscribe", "a208002a000161000162"},
	{"unsuback", "b002002a"},
	{"pingreq", "c000"},
	{"pingresp", "d000"},
	{"disconnect", "e000"},
}

func packetForWriteTest(t *testing.T, wireHex string) (ControlPacket, []byte) {
	t.Helper()
	wire, err := hex.DecodeString(wireHex)
	if err != nil {
		t.Fatal(err)
	}
	reader := bytes.NewReader(wire)
	p, err := ReadPacket(reader)
	if err != nil {
		t.Fatal(err)
	}
	if reader.Len() != 0 {
		t.Fatal("fixture contains trailing bytes")
	}
	// Every packet embeds FixedHeader. Set an intentionally stale length to
	// verify that serialization derives the wire length without changing it.
	reflect.ValueOf(p).Elem().FieldByName("RemainingLength").SetInt(9999)
	return p, wire
}

// TestWriteDoesNotMutatePacket verifies that repeated writes produce the expected
// wire encoding despite a stale RemainingLength and leave all packet data
// unchanged, including when the writer returns an error.
func TestWriteDoesNotMutatePacket(t *testing.T) {
	for _, fixture := range writeFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			p, wire := packetForWriteTest(t, fixture.wire)
			before, _ := packetForWriteTest(t, fixture.wire)
			for i := 0; i < 2; i++ {
				var output bytes.Buffer
				if err := p.Write(&output); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(output.Bytes(), wire) {
					t.Errorf("wire = %x, want %x", output.Bytes(), wire)
				}
				if !reflect.DeepEqual(p, before) {
					t.Errorf("Write modified packet: got %v, want %v", p, before)
				}
			}
			writeErr := errors.New("write failed")
			if err := p.Write(failingPacketWriter{writeErr}); !errors.Is(err, writeErr) {
				t.Errorf("Write error = %v, want %v", err, writeErr)
			}
			if !reflect.DeepEqual(p, before) {
				t.Errorf("failed Write modified packet: got %v, want %v", p, before)
			}
		})
	}
}

type failingPacketWriter struct{ err error }

func (w failingPacketWriter) Write([]byte) (int, error) { return 0, w.err }

// TestWriteConcurrent exercises concurrent writes of a shared packet to separate
// buffers. Run with -race to detect mutations during serialization, such as the
// former updates to FixedHeader.RemainingLength. Without -race, this test checks
// the output and errors but does not reliably detect data races.
// Run with: go test -race ./packets -run '^TestWriteConcurrent$'
func TestWriteConcurrent(t *testing.T) {
	for _, fixture := range writeFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			p, wire := packetForWriteTest(t, fixture.wire)
			var wg sync.WaitGroup
			start := make(chan struct{})
			for i := 0; i < 4; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					for j := 0; j < 100; j++ {
						var output bytes.Buffer
						if err := p.Write(&output); err != nil {
							t.Error(err)
							return
						}
						if !bytes.Equal(output.Bytes(), wire) {
							t.Errorf("wire = %x, want %x", output.Bytes(), wire)
							return
						}
					}
				}()
			}
			close(start)
			wg.Wait()
		})
	}
}
