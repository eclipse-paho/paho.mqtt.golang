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
