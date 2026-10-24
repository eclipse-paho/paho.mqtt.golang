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
 *    Allan Stockdill-Mander
 */

package packets

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"testing/iotest"
)

func TestPacketNames(t *testing.T) {
	if PacketNames[1] != "CONNECT" {
		t.Errorf("PacketNames[1] is %s, should be %s", PacketNames[1], "CONNECT")
	}
	if PacketNames[2] != "CONNACK" {
		t.Errorf("PacketNames[2] is %s, should be %s", PacketNames[2], "CONNACK")
	}
	if PacketNames[3] != "PUBLISH" {
		t.Errorf("PacketNames[3] is %s, should be %s", PacketNames[3], "PUBLISH")
	}
	if PacketNames[4] != "PUBACK" {
		t.Errorf("PacketNames[4] is %s, should be %s", PacketNames[4], "PUBACK")
	}
	if PacketNames[5] != "PUBREC" {
		t.Errorf("PacketNames[5] is %s, should be %s", PacketNames[5], "PUBREC")
	}
	if PacketNames[6] != "PUBREL" {
		t.Errorf("PacketNames[6] is %s, should be %s", PacketNames[6], "PUBREL")
	}
	if PacketNames[7] != "PUBCOMP" {
		t.Errorf("PacketNames[7] is %s, should be %s", PacketNames[7], "PUBCOMP")
	}
	if PacketNames[8] != "SUBSCRIBE" {
		t.Errorf("PacketNames[8] is %s, should be %s", PacketNames[8], "SUBSCRIBE")
	}
	if PacketNames[9] != "SUBACK" {
		t.Errorf("PacketNames[9] is %s, should be %s", PacketNames[9], "SUBACK")
	}
	if PacketNames[10] != "UNSUBSCRIBE" {
		t.Errorf("PacketNames[10] is %s, should be %s", PacketNames[10], "UNSUBSCRIBE")
	}
	if PacketNames[11] != "UNSUBACK" {
		t.Errorf("PacketNames[11] is %s, should be %s", PacketNames[11], "UNSUBACK")
	}
	if PacketNames[12] != "PINGREQ" {
		t.Errorf("PacketNames[12] is %s, should be %s", PacketNames[12], "PINGREQ")
	}
	if PacketNames[13] != "PINGRESP" {
		t.Errorf("PacketNames[13] is %s, should be %s", PacketNames[13], "PINGRESP")
	}
	if PacketNames[14] != "DISCONNECT" {
		t.Errorf("PacketNames[14] is %s, should be %s", PacketNames[14], "DISCONNECT")
	}
}

func TestPacketConsts(t *testing.T) {
	if Connect != 1 {
		t.Errorf("Const for Connect is %d, should be %d", Connect, 1)
	}
	if Connack != 2 {
		t.Errorf("Const for Connack is %d, should be %d", Connack, 2)
	}
	if Publish != 3 {
		t.Errorf("Const for Publish is %d, should be %d", Publish, 3)
	}
	if Puback != 4 {
		t.Errorf("Const for Puback is %d, should be %d", Puback, 4)
	}
	if Pubrec != 5 {
		t.Errorf("Const for Pubrec is %d, should be %d", Pubrec, 5)
	}
	if Pubrel != 6 {
		t.Errorf("Const for Pubrel is %d, should be %d", Pubrel, 6)
	}
	if Pubcomp != 7 {
		t.Errorf("Const for Pubcomp is %d, should be %d", Pubcomp, 7)
	}
	if Subscribe != 8 {
		t.Errorf("Const for Subscribe is %d, should be %d", Subscribe, 8)
	}
	if Suback != 9 {
		t.Errorf("Const for Suback is %d, should be %d", Suback, 9)
	}
	if Unsubscribe != 10 {
		t.Errorf("Const for Unsubscribe is %d, should be %d", Unsubscribe, 10)
	}
	if Unsuback != 11 {
		t.Errorf("Const for Unsuback is %d, should be %d", Unsuback, 11)
	}
	if Pingreq != 12 {
		t.Errorf("Const for Pingreq is %d, should be %d", Pingreq, 12)
	}
	if Pingresp != 13 {
		t.Errorf("Const for Pingresp is %d, should be %d", Pingresp, 13)
	}
	if Disconnect != 14 {
		t.Errorf("Const for Disconnect is %d, should be %d", Disconnect, 14)
	}
}

func TestConnackConsts(t *testing.T) {
	if Accepted != 0x00 {
		t.Errorf("Const for Accepted is %d, should be %d", Accepted, 0)
	}
	if ErrRefusedBadProtocolVersion != 0x01 {
		t.Errorf("Const for RefusedBadProtocolVersion is %d, should be %d", ErrRefusedBadProtocolVersion, 1)
	}
	if ErrRefusedIDRejected != 0x02 {
		t.Errorf("Const for RefusedIDRejected is %d, should be %d", ErrRefusedIDRejected, 2)
	}
	if ErrRefusedServerUnavailable != 0x03 {
		t.Errorf("Const for RefusedServerUnavailable is %d, should be %d", ErrRefusedServerUnavailable, 3)
	}
	if ErrRefusedBadUsernameOrPassword != 0x04 {
		t.Errorf("Const for RefusedBadUsernameOrPassword is %d, should be %d", ErrRefusedBadUsernameOrPassword, 4)
	}
	if ErrRefusedNotAuthorised != 0x05 {
		t.Errorf("Const for RefusedNotAuthorised is %d, should be %d", ErrRefusedNotAuthorised, 5)
	}
}

func TestConnectPacket(t *testing.T) {
	connectPacketBytes := bytes.NewBuffer([]byte{16, 52, 0, 4, 77, 81, 84, 84, 4, 204, 0, 0, 0, 0, 0, 4, 116, 101, 115, 116, 0, 12, 84, 101, 115, 116, 32, 80, 97, 121, 108, 111, 97, 100, 0, 8, 116, 101, 115, 116, 117, 115, 101, 114, 0, 8, 116, 101, 115, 116, 112, 97, 115, 115})
	packet, err := ReadPacket(connectPacketBytes)
	if err != nil {
		t.Fatalf("Error reading packet: %s", err.Error())
	}
	cp := packet.(*ConnectPacket)
	if cp.ProtocolName != "MQTT" {
		t.Errorf("Connect Packet ProtocolName is %s, should be %s", cp.ProtocolName, "MQTT")
	}
	if cp.ProtocolVersion != 4 {
		t.Errorf("Connect Packet ProtocolVersion is %d, should be %d", cp.ProtocolVersion, 4)
	}
	if cp.UsernameFlag != true {
		t.Errorf("Connect Packet UsernameFlag is %t, should be %t", cp.UsernameFlag, true)
	}
	if cp.Username != "testuser" {
		t.Errorf("Connect Packet Username is %s, should be %s", cp.Username, "testuser")
	}
	if cp.PasswordFlag != true {
		t.Errorf("Connect Packet PasswordFlag is %t, should be %t", cp.PasswordFlag, true)
	}
	if string(cp.Password) != "testpass" {
		t.Errorf("Connect Packet Password is %s, should be %s", string(cp.Password), "testpass")
	}
	if cp.WillFlag != true {
		t.Errorf("Connect Packet WillFlag is %t, should be %t", cp.WillFlag, true)
	}
	if cp.WillTopic != "test" {
		t.Errorf("Connect Packet WillTopic is %s, should be %s", cp.WillTopic, "test")
	}
	if cp.WillQos != 1 {
		t.Errorf("Connect Packet WillQos is %d, should be %d", cp.WillQos, 1)
	}
	if cp.WillRetain != false {
		t.Errorf("Connect Packet WillRetain is %t, should be %t", cp.WillRetain, false)
	}
	if string(cp.WillMessage) != "Test Payload" {
		t.Errorf("Connect Packet WillMessage is %s, should be %s", string(cp.WillMessage), "Test Payload")
	}
}

func TestReadPacketWithLimitRejectsOversizedPacket(t *testing.T) {
	// A maximum MQTT Remaining Length encoded in only four bytes must be
	// rejected without attempting to allocate its 256 MiB payload.
	packetHeader := []byte{0x30, 0xff, 0xff, 0xff, 0x7f}

	packet, err := ReadPacketWithLimit(bytes.NewReader(packetHeader), 1024)
	if packet != nil {
		t.Fatalf("expected no packet, got %T", packet)
	}
	if !errors.Is(err, ErrPacketTooLarge) {
		t.Fatalf("expected ErrPacketTooLarge, got %v", err)
	}
}

func TestReadPacketWithLimitAcceptsPacketWithinLimit(t *testing.T) {
	packet, err := ReadPacketWithLimit(bytes.NewReader([]byte{0x20, 0x02, 0x00, 0x00}), 2)
	if err != nil {
		t.Fatalf("expected packet within limit to be accepted, got %v", err)
	}
	if _, ok := packet.(*ConnackPacket); !ok {
		t.Fatalf("expected ConnackPacket, got %T", packet)
	}
}

func TestReadPacketMessageID(t *testing.T) {
	for _, header := range []byte{0x40, 0x50, 0x62, 0x70, 0x90, 0xb0} {
		t.Run(PacketNames[header>>4], func(t *testing.T) {
			t.Run("truncated", func(t *testing.T) {
				// The body matches Remaining Length, but cannot contain a full Message ID.
				_, err := ReadPacket(bytes.NewReader([]byte{header, 0x01, 0x02}))
				if !errors.Is(err, io.ErrUnexpectedEOF) {
					t.Fatalf("expected io.ErrUnexpectedEOF, got %v", err)
				}
			})
			t.Run("complete", func(t *testing.T) {
				wire := []byte{header, 0x02, 0x12, 0x34}
				if header>>4 == Suback {
					wire[1]++
					wire = append(wire, 0x02)
				}
				packet, err := ReadPacket(bytes.NewReader(wire))
				if err != nil {
					t.Fatalf("error reading complete packet: %v", err)
				}
				if got := packet.Details().MessageID; got != 0x1234 {
					t.Errorf("Message ID = %#x, want 0x1234", got)
				}
				if sa, ok := packet.(*SubackPacket); ok && !bytes.Equal(sa.ReturnCodes, []byte{0x02}) {
					t.Errorf("return codes = %v, want [2]", sa.ReturnCodes)
				}
			})
		})
	}
}

func TestReadPacketClientIdentifier(t *testing.T) {
	for _, tt := range []struct {
		name    string
		field   []byte
		want    string
		wantErr error
	}{
		{name: "complete", field: []byte{0, 2, 'i', 'd'}, want: "id"},
		{name: "empty", field: []byte{0, 0}},
		{name: "missing body", field: []byte{0, 2}, wantErr: io.EOF},
		{name: "truncated body", field: []byte{0, 2, 'i'}, wantErr: io.ErrUnexpectedEOF},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// MQTT 3.1.1 CONNECT with Clean Session, allowing an empty client ID.
			body := append([]byte{0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 0}, tt.field...)
			// Remaining Length is correct even when the field's own length is not.
			wire := append([]byte{0x10, byte(len(body))}, body...)
			packet, err := ReadPacket(bytes.NewReader(wire))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			connect, ok := packet.(*ConnectPacket)
			if !ok {
				t.Fatalf("packet = %T, want *ConnectPacket", packet)
			}
			if connect.ClientIdentifier != tt.want {
				t.Errorf("client identifier = %q, want %q", connect.ClientIdentifier, tt.want)
			}
		})
	}
}

func TestSubackReturnCodes(t *testing.T) {
	check := func(t *testing.T, packet *SubackPacket, err error, codes []byte, wantError bool) {
		t.Helper()
		if wantError {
			if !errors.Is(err, ErrMalformedSuback) {
				t.Errorf("error = %v, want ErrMalformedSuback", err)
			}
			if packet != nil && len(packet.ReturnCodes) != 0 {
				t.Errorf("invalid packet populated return codes: %v", packet.ReturnCodes)
			}
			return
		}
		if err != nil {
			t.Fatalf("error decoding valid return codes: %v", err)
		}
		if packet == nil {
			t.Fatal("expected a SUBACK packet")
		}
		if packet.MessageID != 0x1234 {
			t.Errorf("message ID = %#x, want 0x1234", packet.MessageID)
		}
		if !bytes.Equal(packet.ReturnCodes, codes) {
			t.Errorf("return codes = %v, want %v", packet.ReturnCodes, codes)
		}
	}
	for _, tt := range []struct {
		name      string
		codes     []byte
		wantError bool
	}{
		{name: "qos0", codes: []byte{0}},
		{name: "qos1", codes: []byte{1}},
		{name: "qos2", codes: []byte{2}},
		{name: "failure", codes: []byte{0x80}},
		{name: "mixed valid", codes: []byte{0, 1, 2, 0x80}},
		{name: "reserved 03", codes: []byte{3}, wantError: true},
		{name: "reserved 04", codes: []byte{4}, wantError: true},
		{name: "reserved 7f", codes: []byte{0x7f}, wantError: true},
		{name: "reserved 81", codes: []byte{0x81}, wantError: true},
		{name: "reserved fe", codes: []byte{0xfe}, wantError: true},
		{name: "reserved ff", codes: []byte{0xff}, wantError: true},
		{name: "invalid after valid", codes: []byte{0, 1, 2, 0xfe}, wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			body := append([]byte{0x12, 0x34}, tt.codes...)
			t.Run("Unpack", func(t *testing.T) {
				packet := NewControlPacket(Suback).(*SubackPacket)
				err := packet.Unpack(bytes.NewReader(body))
				check(t, packet, err, tt.codes, tt.wantError)
			})
			t.Run("ReadPacket", func(t *testing.T) {
				wire := append([]byte{0x90, byte(len(body))}, body...)
				wire = append(wire, 0xd0, 0) // Following PINGRESP must remain unread.
				reader := bytes.NewReader(wire)
				packet, err := ReadPacket(reader)
				suback, _ := packet.(*SubackPacket)
				check(t, suback, err, tt.codes, tt.wantError)
				next, err := ReadPacket(reader)
				if err != nil {
					t.Fatalf("error reading following packet: %v", err)
				}
				if _, ok := next.(*PingrespPacket); !ok {
					t.Errorf("following packet = %T, want *PingrespPacket", next)
				}
			})
		})
	}
}

func TestPackUnpackControlPackets(t *testing.T) {
	// The UNSUBSCRIBE payload requires at least one topic filter (MQTT-3.10.3-2),
	// so the round-trip fixture cannot use the zero-value packet with no topics.
	unsubscribe := NewControlPacket(Unsubscribe).(*UnsubscribePacket)
	unsubscribe.MessageID = 0x1234
	unsubscribe.Topics = []string{"a/b"}
	packets := []ControlPacket{
		NewControlPacket(Connect).(*ConnectPacket),
		NewControlPacket(Connack).(*ConnackPacket),
		NewControlPacket(Publish).(*PublishPacket),
		NewControlPacket(Puback).(*PubackPacket),
		NewControlPacket(Pubrec).(*PubrecPacket),
		NewControlPacket(Pubrel).(*PubrelPacket),
		NewControlPacket(Pubcomp).(*PubcompPacket),
		NewControlPacket(Subscribe).(*SubscribePacket),
		NewControlPacket(Suback).(*SubackPacket),
		unsubscribe,
		NewControlPacket(Unsuback).(*UnsubackPacket),
		NewControlPacket(Pingreq).(*PingreqPacket),
		NewControlPacket(Pingresp).(*PingrespPacket),
		NewControlPacket(Disconnect).(*DisconnectPacket),
	}
	buf := new(bytes.Buffer)
	for _, packet := range packets {
		buf.Reset()
		if err := packet.Write(buf); err != nil {
			t.Errorf("Write of %T returned error: %s", packet, err)
		}
		wire := append([]byte(nil), buf.Bytes()...)
		read, err := ReadPacket(buf)
		if err != nil {
			t.Errorf("Read of packed %T returned error: %s", packet, err)
		}
		// RemainingLength is populated by decoding, not by Write. Compare the
		// serialized packets rather than their diagnostic strings.
		buf.Reset()
		if err := read.Write(buf); err != nil {
			t.Fatalf("Write of decoded %T returned error: %s", read, err)
		}
		if !bytes.Equal(buf.Bytes(), wire) {
			t.Errorf("Read of packed %T changed wire encoding: got %x, want %x", packet, buf.Bytes(), wire)
		}
	}
}

func TestDecodeUint16(t *testing.T) {
	readErr := errors.New("read failed")
	for _, tt := range []struct {
		name    string
		reader  io.Reader
		want    uint16
		wantErr error
	}{
		{name: "complete", reader: bytes.NewReader([]byte{0x12, 0x34}), want: 0x1234},
		{name: "fragmented", reader: iotest.OneByteReader(bytes.NewReader([]byte{0x12, 0x34})), want: 0x1234},
		{name: "complete with EOF", reader: iotest.DataErrReader(bytes.NewReader([]byte{0x12, 0x34})), want: 0x1234},
		{name: "empty", reader: bytes.NewReader(nil), wantErr: io.EOF},
		{name: "truncated", reader: bytes.NewReader([]byte{0x12}), wantErr: io.ErrUnexpectedEOF},
		{name: "read error", reader: iotest.ErrReader(readErr), wantErr: readErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeUint16(tt.reader)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("value = %#x, want %#x", got, tt.want)
			}
		})
	}
}

// emptyFirstReader returns no data on its first read, without indicating EOF.
type emptyFirstReader struct {
	io.Reader
	emptyRead bool
}

func (r *emptyFirstReader) Read(p []byte) (int, error) {
	if !r.emptyRead {
		r.emptyRead = true
		return 0, nil
	}
	return r.Reader.Read(p)
}

func TestDecodeByte(t *testing.T) {
	readErr := errors.New("read failed")
	for _, tt := range []struct {
		name      string
		reader    io.Reader
		want      byte
		wantErr   error
		remaining []byte
	}{
		{name: "complete", reader: bytes.NewReader([]byte{0x56}), want: 0x56},
		{name: "empty read then byte", reader: &emptyFirstReader{Reader: bytes.NewReader([]byte{0x56})}, want: 0x56},
		{name: "complete with EOF", reader: iotest.DataErrReader(bytes.NewReader([]byte{0x56})), want: 0x56},
		{name: "missing", reader: bytes.NewReader(nil), wantErr: io.EOF},
		{name: "read error", reader: iotest.ErrReader(readErr), wantErr: readErr},
		{name: "following byte", reader: bytes.NewReader([]byte{0x56, 0x78}), want: 0x56, remaining: []byte{0x78}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeByte(tt.reader)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("value = %#x, want %#x", got, tt.want)
			}
			if tt.wantErr == nil {
				remaining, err := io.ReadAll(tt.reader)
				if err != nil || !bytes.Equal(remaining, tt.remaining) {
					t.Errorf("remaining bytes = %v, error = %v, want %v", remaining, err, tt.remaining)
				}
			}
		})
	}
}

func TestDecodeLengthPrefixedFields(t *testing.T) {
	readErr := errors.New("read failed")
	for _, decoder := range []struct {
		name   string
		decode func(io.Reader) ([]byte, error)
	}{
		{name: "bytes", decode: decodeBytes},
		{name: "string", decode: func(r io.Reader) ([]byte, error) {
			value, err := decodeString(r)
			return []byte(value), err
		}},
	} {
		t.Run(decoder.name, func(t *testing.T) {
			for _, tt := range []struct {
				name      string
				reader    io.Reader
				want      []byte
				wantErr   error
				remaining []byte
			}{
				{name: "complete", reader: bytes.NewReader([]byte{0, 3, 'a', 'b', 'c'}), want: []byte("abc")},
				{name: "fragmented", reader: iotest.OneByteReader(bytes.NewReader([]byte{0, 3, 'a', 'b', 'c'})), want: []byte("abc")},
				{name: "complete with EOF", reader: iotest.DataErrReader(bytes.NewReader([]byte{0, 3, 'a', 'b', 'c'})), want: []byte("abc")},
				{name: "empty terminal field", reader: bytes.NewReader([]byte{0, 0})},
				{name: "missing length", reader: bytes.NewReader(nil), wantErr: io.EOF},
				{name: "partial length", reader: bytes.NewReader([]byte{0}), wantErr: io.ErrUnexpectedEOF},
				{name: "missing body", reader: bytes.NewReader([]byte{0, 3}), wantErr: io.EOF},
				{name: "partial body", reader: bytes.NewReader([]byte{0, 3, 'a'}), wantErr: io.ErrUnexpectedEOF},
				{name: "fragmented partial body", reader: iotest.OneByteReader(bytes.NewReader([]byte{0, 3, 'a', 'b'})), wantErr: io.ErrUnexpectedEOF},
				{name: "length read error", reader: iotest.ErrReader(readErr), wantErr: readErr},
				{name: "body read error", reader: io.MultiReader(bytes.NewReader([]byte{0, 3}), iotest.ErrReader(readErr)), wantErr: readErr},
				{name: "partial body read error", reader: io.MultiReader(bytes.NewReader([]byte{0, 3, 'a'}), iotest.ErrReader(readErr)), wantErr: readErr},
				{name: "following field", reader: bytes.NewReader([]byte{0, 3, 'a', 'b', 'c', 0, 1, 'd'}), want: []byte("abc"), remaining: []byte{0, 1, 'd'}},
				{name: "empty before following field", reader: bytes.NewReader([]byte{0, 0, 0, 1, 'd'}), remaining: []byte{0, 1, 'd'}},
			} {
				t.Run(tt.name, func(t *testing.T) {
					got, err := decoder.decode(tt.reader)
					if !errors.Is(err, tt.wantErr) {
						t.Errorf("error = %v, want %v", err, tt.wantErr)
					}
					if !bytes.Equal(got, tt.want) {
						t.Errorf("value = %v, want %v", got, tt.want)
					}
					if tt.wantErr == nil {
						remaining, err := io.ReadAll(tt.reader)
						if err != nil || !bytes.Equal(remaining, tt.remaining) {
							t.Errorf("remaining bytes = %v, error = %v, want %v", remaining, err, tt.remaining)
						}
					}
				})
			}
		})
	}
}

func TestEncoding(t *testing.T) {
	if res, err := decodeByte(bytes.NewBuffer([]byte{0x56})); res != 0x56 || err != nil {
		t.Errorf("decodeByte([0x56]) did not return (0x56, nil) but (0x%X, %v)", res, err)
	}
	if res, err := decodeUint16(bytes.NewBuffer([]byte{0x56, 0x78})); res != 22136 || err != nil {
		t.Errorf("decodeUint16([0x5678]) did not return (22136, nil) but (%d, %v)", res, err)
	}
	if res := encodeUint16(22136); !bytes.Equal(res, []byte{0x56, 0x78}) {
		t.Errorf("encodeUint16(22136) did not return [0x5678] but [0x%X]", res)
	}

	strings := map[string][]byte{
		"foo":         {0x00, 0x03, 'f', 'o', 'o'},
		"\U0000FEFF":  {0x00, 0x03, 0xEF, 0xBB, 0xBF},
		"A\U0002A6D4": {0x00, 0x05, 'A', 0xF0, 0xAA, 0x9B, 0x94},
	}
	for str, encoded := range strings {
		if res, err := decodeString(bytes.NewBuffer(encoded)); res != str || err != nil {
			t.Errorf("decodeString(%v) did not return (%q, nil), but (%q, %v)", encoded, str, res, err)
		}
		if res := encodeString(str); !bytes.Equal(res, encoded) {
			t.Errorf("encodeString(%q) did not return [0x%X], but [0x%X]", str, encoded, res)
		}
	}

	lengths := map[int][]byte{
		0:         {0x00},
		127:       {0x7F},
		128:       {0x80, 0x01},
		16383:     {0xFF, 0x7F},
		16384:     {0x80, 0x80, 0x01},
		2097151:   {0xFF, 0xFF, 0x7F},
		2097152:   {0x80, 0x80, 0x80, 0x01},
		268435455: {0xFF, 0xFF, 0xFF, 0x7F},
	}
	for length, encoded := range lengths {
		if res, err := decodeLength(bytes.NewBuffer(encoded)); res != length || err != nil {
			t.Errorf("decodeLength([0x%X]) did not return (%d, nil) but (%d, %v)", encoded, length, res, err)
		}
		if res, err := encodeLength(length); !bytes.Equal(res, encoded) || err != nil {
			t.Errorf("encodeLength(%d) did not return [0x%X], but [0x%X]", length, encoded, res)
		}
	}

	// Encoding or decoding data longer than 268,435,455 bytes should fail with an error (this check was added to
	// the 3.1.1 spec after publication). Checking this when sending avoids sending invalid packet.
	tooLong := []byte{0xFF, 0xFF, 0xFF, 0x80, 0x01}
	if _, err := decodeLength(bytes.NewBuffer(tooLong)); err == nil {
		t.Errorf("decodeLength([0x%X]) did not return error", tooLong)
	}
	if _, err := encodeLength(268435456); err == nil {
		t.Error("encodeLength(268435456) did not return error")
	}

	// When encoding a string longer than 2^16 bytes, the result must not exceed the length of the 16-bit header.
	// Previously this was possible due to the use of uint16(len(field)) for the length and then writing the entirety of
	// field.
	overlengthStr := bytes.Repeat([]byte("A"), 65600)         // longer than 2^16
	if res := encodeBytes(overlengthStr); len(res) != 65537 { // Two byte length so 65535 + 2 = 65537
		t.Errorf("encodeBytes did not truncate overlength data (expected len 65538, received %d", len(res))
	}

}

// TestPacketBodyLength checks that a packet whose Remaining Length does not match
// the body its type defines is rejected (#778). Truncated bodies are already
// covered by TestReadPacketMessageID.
func TestPacketBodyLength(t *testing.T) {
	for _, tt := range []struct {
		name   string
		header byte
		body   []byte
	}{
		{name: "connack", header: 0x20, body: []byte{0x00, 0x00}},
		{name: "puback", header: 0x40, body: []byte{0x12, 0x34}},
		{name: "pubrec", header: 0x50, body: []byte{0x12, 0x34}},
		{name: "pubrel", header: 0x62, body: []byte{0x12, 0x34}},
		{name: "pubcomp", header: 0x70, body: []byte{0x12, 0x34}},
		{name: "unsuback", header: 0xb0, body: []byte{0x12, 0x34}},
		{name: "pingreq", header: 0xc0},
		{name: "pingresp", header: 0xd0},
		{name: "disconnect", header: 0xe0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			packetType := tt.header >> 4

			t.Run("standard length", func(t *testing.T) {
				wire := append([]byte{tt.header, byte(len(tt.body))}, tt.body...)
				if _, err := ReadPacket(bytes.NewReader(wire)); err != nil {
					t.Errorf("error reading packet with its standard body length: %v", err)
				}
				if err := NewControlPacket(packetType).Unpack(bytes.NewReader(tt.body)); err != nil {
					t.Errorf("error unpacking packet with its standard body length: %v", err)
				}
			})

			extra := append(append([]byte(nil), tt.body...), 0xff)

			t.Run("trailing byte", func(t *testing.T) {
				// Unpack consumes the packet's fields and leaves any unread bytes
				// in the reader untouched, so it stays safe to call on a stream.
				// Rejecting the leftover bytes is the packet reader's job.
				packet := NewControlPacket(packetType)
				if err := packet.Unpack(bytes.NewReader(extra)); err != nil {
					t.Errorf("error = %v, want nil (the packet reader rejects unread body bytes)", err)
				}
			})

			t.Run("declared length", func(t *testing.T) {
				// Overstating the length of a fixed-length body is rejected before
				// the declared bytes are read, so a hostile peer cannot make the
				// reader allocate for a packet that can never be valid.
				wire := append([]byte{tt.header, byte(len(extra))}, extra...)
				packet, err := ReadPacket(bytes.NewReader(wire))
				if !errors.Is(err, ErrUnexpectedBodyData) {
					t.Fatalf("error = %v, want ErrUnexpectedBodyData", err)
				}
				if packet != nil {
					t.Errorf("packet = %T, want nil so that it cannot be processed", packet)
				}
			})
		})
	}
}

// TestVariableLengthBodyBoundaries covers the packets whose body size depends on
// their fields (#778): the payload of a PUBLISH, the return codes of a SUBACK and
// the topic filters of CONNECT/SUBSCRIBE/UNSUBSCRIBE must not be mistaken for
// trailing data, while bytes that no field accounts for must be rejected.
func TestVariableLengthBodyBoundaries(t *testing.T) {
	t.Run("publish", func(t *testing.T) {
		wire := []byte{0x30, 8, 0, 1, 'a', 'h', 'e', 'l', 'l', 'o'}
		packet, err := ReadPacket(bytes.NewReader(wire))
		if err != nil {
			t.Fatalf("error reading publish: %v", err)
		}
		pub, ok := packet.(*PublishPacket)
		if !ok {
			t.Fatalf("packet = %T, want *PublishPacket", packet)
		}
		if string(pub.Payload) != "hello" {
			t.Errorf("payload = %q, want \"hello\"", pub.Payload)
		}
	})

	t.Run("publish fragmented body", func(t *testing.T) {
		pub := NewControlPacket(Publish).(*PublishPacket)
		pub.FixedHeader.RemainingLength = 8
		body := []byte{0, 1, 'a', 'h', 'e', 'l', 'l', 'o'}

		if err := pub.Unpack(iotest.OneByteReader(bytes.NewReader(body))); err != nil {
			t.Fatalf("error unpacking publish: %v", err)
		}
		if string(pub.Payload) != "hello" {
			t.Errorf("payload = %q, want \"hello\"", pub.Payload)
		}
	})

	t.Run("suback return codes", func(t *testing.T) {
		wire := []byte{0x90, 5, 0x12, 0x34, 0, 1, 0x80}
		packet, err := ReadPacket(bytes.NewReader(wire))
		if err != nil {
			t.Fatalf("error reading suback: %v", err)
		}
		sa, ok := packet.(*SubackPacket)
		if !ok {
			t.Fatalf("packet = %T, want *SubackPacket", packet)
		}
		if !bytes.Equal(sa.ReturnCodes, []byte{0, 1, 0x80}) {
			t.Errorf("return codes = %v, want [0 1 128]", sa.ReturnCodes)
		}
	})

	t.Run("connect trailing byte", func(t *testing.T) {
		// MQTT 3.1.1 CONNECT with clean session and client ID "id", plus a byte
		// that no CONNECT field accounts for.
		body := []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 0, 0, 2, 'i', 'd', 0xff}
		wire := append([]byte{0x10, byte(len(body))}, body...)

		if _, err := ReadPacket(bytes.NewReader(wire)); !errors.Is(err, ErrUnexpectedBodyData) {
			t.Errorf("error = %v, want ErrUnexpectedBodyData", err)
		}
	})

	t.Run("malformed then valid", func(t *testing.T) {
		// CONNECT's declared length is consistent with its fields, so its body is
		// read in full before the trailing byte is rejected. The packet after it
		// must still be framed correctly rather than parsed as garbage.
		body := []byte{0, 4, 'M', 'Q', 'T', 'T', 4, 2, 0, 0, 0, 2, 'i', 'd', 0xff}
		reader := bytes.NewReader(append(append([]byte{0x10, byte(len(body))}, body...), 0xd0, 0))

		if _, err := ReadPacket(reader); !errors.Is(err, ErrUnexpectedBodyData) {
			t.Fatalf("error = %v, want ErrUnexpectedBodyData", err)
		}
		next, err := ReadPacket(reader)
		if err != nil {
			t.Fatalf("error reading following packet: %v", err)
		}
		if _, ok := next.(*PingrespPacket); !ok {
			t.Errorf("following packet = %T, want *PingrespPacket", next)
		}
	})

	t.Run("subscribe trailing byte", func(t *testing.T) {
		// Message ID, the filter "a" with its QoS byte, then a stray byte. The
		// filter loop is driven by Remaining Length, so the stray byte is parsed
		// as the length prefix of another filter and the truncation is reported.
		body := []byte{0x12, 0x34, 0, 1, 'a', 0, 0xff}
		wire := append([]byte{0x82, byte(len(body))}, body...)

		if _, err := ReadPacket(bytes.NewReader(wire)); !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("error = %v, want io.ErrUnexpectedEOF", err)
		}
	})

	t.Run("unsubscribe", func(t *testing.T) {
		t.Run("topic filters", func(t *testing.T) {
			body := []byte{0x12, 0x34, 0, 1, 'a', 0, 2, 'b', 'c'}
			wire := append([]byte{0xa2, byte(len(body))}, body...)

			packet, err := ReadPacket(bytes.NewReader(wire))
			if err != nil {
				t.Fatalf("error reading unsubscribe: %v", err)
			}
			un, ok := packet.(*UnsubscribePacket)
			if !ok {
				t.Fatalf("packet = %T, want *UnsubscribePacket", packet)
			}
			if len(un.Topics) != 2 || un.Topics[0] != "a" || un.Topics[1] != "bc" {
				t.Errorf("topics = %q, want [a bc]", un.Topics)
			}
		})

		t.Run("no topic filters", func(t *testing.T) {
			// MQTT-3.10.3-2: the payload must contain at least one topic filter, so
			// a body holding only the packet identifier is malformed.
			body := []byte{0x12, 0x34}
			wire := append([]byte{0xa2, byte(len(body))}, body...)

			if _, err := ReadPacket(bytes.NewReader(wire)); !errors.Is(err, ErrMalformedUnsubscribe) {
				t.Errorf("error = %v, want ErrMalformedUnsubscribe", err)
			}
		})

		t.Run("empty topic filter", func(t *testing.T) {
			// The filters run until the end of the body; an empty one is not a
			// topic filter, so it cannot mark the end of the payload.
			body := []byte{0x12, 0x34, 0, 1, 'a', 0, 0}
			wire := append([]byte{0xa2, byte(len(body))}, body...)

			if _, err := ReadPacket(bytes.NewReader(wire)); !errors.Is(err, ErrMalformedUnsubscribe) {
				t.Errorf("error = %v, want ErrMalformedUnsubscribe", err)
			}
		})

		t.Run("truncated topic filter", func(t *testing.T) {
			// A filter whose declared length runs past the end of the body must be
			// reported; the unpacker used to discard the decode error and accept an
			// UNSUBSCRIBE with no filters at all.
			body := []byte{0x12, 0x34, 0, 2, 'a'}
			wire := append([]byte{0xa2, byte(len(body))}, body...)

			if _, err := ReadPacket(bytes.NewReader(wire)); !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Errorf("error = %v, want io.ErrUnexpectedEOF", err)
			}
		})
	})
}

func TestReadPacketRejectsUnreadBodyBytes(t *testing.T) {
	// A CONNECT body whose fields end after the empty client id: protocol name
	// "MQTT", level 4, clean-session flag, zero keepalive. The final byte is
	// declared in the Remaining Length but never consumed by Unpack, so the
	// packet reader must reject it instead of accepting the trailing data.
	body := []byte{0x00, 0x04, 'M', 'Q', 'T', 'T', 0x04, 0x02, 0x00, 0x00, 0x00, 0x00, 0xff}
	wire := append([]byte{0x10, byte(len(body))}, body...)

	if _, err := ReadPacket(bytes.NewReader(wire)); !errors.Is(err, ErrUnexpectedBodyData) {
		t.Errorf("error = %v, want ErrUnexpectedBodyData", err)
	}
}
