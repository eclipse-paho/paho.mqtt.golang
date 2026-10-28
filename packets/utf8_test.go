package packets

import (
	"bytes"
	"testing"
)

func TestDecodeStringRejectsInvalidUTF8(t *testing.T) {
	cases := []struct {
		name    string
		encoded []byte // 2-byte length prefix + field bytes
		wantErr bool
	}{
		{"valid multibyte", []byte{0x00, 0x03, 'h', 0xC3, 0xA9}, false},
		{"U+FEFF is preserved, not stripped", []byte{0x00, 0x03, 0xEF, 0xBB, 0xBF}, false},
		{"null character U+0000", []byte{0x00, 0x01, 0x00}, true},
		{"overlong encoding of U+0040", []byte{0x00, 0x02, 0xC0, 0x80}, true},
		{"surrogate half U+D800", []byte{0x00, 0x03, 0xED, 0xA0, 0x80}, true},
		{"truncated multibyte sequence", []byte{0x00, 0x02, 0xE4, 0xB8}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := decodeString(bytes.NewReader(c.encoded))
			if c.wantErr {
				if err == nil {
					t.Errorf("decodeString(%x) = %q, nil; want error", c.encoded, got)
				}
				return
			}
			if err != nil {
				t.Errorf("decodeString(%x) returned unexpected error: %v", c.encoded, err)
			}
		})
	}
}

func TestEncodeStringCheckedRejectsInvalid(t *testing.T) {
	cases := []struct {
		name    string
		field   string
		wantErr bool
	}{
		{"valid ascii", "a/b/c", false},
		{"valid multibyte", "h\xc3\xa9llo/\xe4\xb8\xad\xe6\x96\x87", false},
		{"U+FEFF preserved", "\uFEFF/topic", false},
		{"null character", "bad\x00topic", true},
		{"invalid utf-8 byte sequence", string([]byte{0xC0, 0x80}), true},
		{"surrogate encoding", string([]byte{0xED, 0xA0, 0x80}), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := encodeStringChecked(c.field)
			if c.wantErr {
				if err == nil {
					t.Errorf("encodeStringChecked(%q) = %x, nil; want error", c.field, out)
				}
				return
			}
			if err != nil {
				t.Errorf("encodeStringChecked(%q) returned unexpected error: %v", c.field, err)
			}
		})
	}
}

func TestPacketWritesRejectInvalidStringFields(t *testing.T) {
	surrogate := string([]byte{0xED, 0xA0, 0x80})
	if err := (&PublishPacket{TopicName: "bad\x00topic"}).Write(&bytes.Buffer{}); err == nil {
		t.Error("PublishPacket.Write with U+0000 in TopicName: want error, got nil")
	}
	if err := (&PublishPacket{TopicName: surrogate}).Write(&bytes.Buffer{}); err == nil {
		t.Error("PublishPacket.Write with a surrogate-encoded TopicName: want error, got nil")
	}
	if err := (&ConnectPacket{ClientIdentifier: surrogate}).Write(&bytes.Buffer{}); err == nil {
		t.Error("ConnectPacket.Write with a surrogate-encoded ClientIdentifier: want error, got nil")
	}
	if err := (&SubscribePacket{Topics: []string{"ok", "bad\x00"}}).Write(&bytes.Buffer{}); err == nil {
		t.Error("SubscribePacket.Write with U+0000 in a topic filter: want error, got nil")
	}
	if err := (&UnsubscribePacket{Topics: []string{"bad\x00"}}).Write(&bytes.Buffer{}); err == nil {
		t.Error("UnsubscribePacket.Write with U+0000 in a topic filter: want error, got nil")
	}
	// An empty ClientIdentifier is legitimate (the broker assigns one).
	if err := (&ConnectPacket{}).Write(&bytes.Buffer{}); err != nil {
		t.Errorf("ConnectPacket.Write with an empty ClientIdentifier returned an error: %v", err)
	}
}

func TestPublishRoundTripPreservesMultibyteTopic(t *testing.T) {
	topic := "h\xc3\xa9llo/\xe4\xb8\xad\xe6\x96\x87/\uFEFF/#"
	p := &PublishPacket{FixedHeader: FixedHeader{MessageType: Publish}, TopicName: topic, Payload: []byte("payload")}
	var buf bytes.Buffer
	if err := p.Write(&buf); err != nil {
		t.Fatalf("Write returned unexpected error: %v", err)
	}
	cp, err := ReadPacket(&buf)
	if err != nil {
		t.Fatalf("ReadPacket returned unexpected error: %v", err)
	}
	got, ok := cp.(*PublishPacket)
	if !ok {
		t.Fatalf("ReadPacket returned a %T, expected a *PublishPacket", cp)
	}
	if got.TopicName != topic {
		t.Errorf("topic name did not survive the round trip: got %q, want %q", got.TopicName, topic)
	}
}

func TestPacketValidateRejectsInvalidStringFields(t *testing.T) {
	surrogate := string([]byte{0xED, 0xA0, 0x80})
	valid := "valid/topic"
	if err := (&PublishPacket{TopicName: valid}).Validate(); err != nil {
		t.Errorf("PublishPacket.Validate(%q) returned unexpected error: %v", valid, err)
	}
	for _, bad := range []string{"bad\x00topic", surrogate} {
		if err := (&PublishPacket{TopicName: bad}).Validate(); err == nil {
			t.Errorf("PublishPacket.Validate(%q): want error, got nil", bad)
		}
		if err := (&SubscribePacket{Topics: []string{valid, bad}}).Validate(); err == nil {
			t.Errorf("SubscribePacket.Validate with topic %q: want error, got nil", bad)
		}
		if err := (&UnsubscribePacket{Topics: []string{bad}}).Validate(); err == nil {
			t.Errorf("UnsubscribePacket.Validate with topic %q: want error, got nil", bad)
		}
	}
}

// TestUnsubscribeUnpackRejectsInvalidTopics covers inbound UNSUBSCRIBE packets:
// a broker-supplied topic filter must be a valid MQTT UTF-8 string, even though
// it never passes through a Write path (issue #780).
func TestUnsubscribeUnpackRejectsInvalidTopics(t *testing.T) {
	cases := []struct {
		name    string
		filters [][]byte
		wantErr bool
	}{
		{"valid ascii and multibyte", [][]byte{[]byte("a/b"), []byte("c/\xc3\xa9")}, false},
		{"U+FEFF is preserved", [][]byte{[]byte("\xEF\xBB\xBF/topic")}, false},
		{"null byte in topic", [][]byte{[]byte("bad\x00topic")}, true},
		{"surrogate half", [][]byte{{0xED, 0xA0, 0x80}}, true},
		{"overlong encoding", [][]byte{{0xC0, 0x80}}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body := bytes.NewBuffer(nil)
			body.Write(encodeUint16(1)) // message ID
			for _, f := range c.filters {
				body.Write(encodeBytes(f))
			}
			head, err := (&FixedHeader{MessageType: Unsubscribe, Qos: 1}).pack(body.Len())
			if err != nil {
				t.Fatalf("fixed header pack failed: %v", err)
			}
			head.Write(body.Bytes())
			cp, err := ReadPacket(&head)
			if c.wantErr {
				if err == nil {
					t.Errorf("ReadPacket with filter %q: want error, got %T", c.filters, cp)
				}
				return
			}
			if err != nil {
				t.Fatalf("ReadPacket returned unexpected error: %v", err)
			}
			got, ok := cp.(*UnsubscribePacket)
			if !ok {
				t.Fatalf("ReadPacket returned a %T, expected an *UnsubscribePacket", cp)
			}
			if len(got.Topics) != len(c.filters) {
				t.Fatalf("decoded %d topics, want %d: %q", len(got.Topics), len(c.filters), got.Topics)
			}
			for i, f := range c.filters {
				if got.Topics[i] != string(f) {
					t.Errorf("topic %d did not survive the round trip: got %q, want %q", i, got.Topics[i], string(f))
				}
			}
		})
	}
}
