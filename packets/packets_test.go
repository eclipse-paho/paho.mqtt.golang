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
	"fmt"
	"reflect"
	"testing"
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

func TestPackUnpackControlPackets(t *testing.T) {
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
		NewControlPacket(Unsubscribe).(*UnsubscribePacket),
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
		read, err := ReadPacket(buf)
		if err != nil {
			t.Errorf("Read of packed %T returned error: %s", packet, err)
		}
		if read.String() != packet.String() {
			t.Errorf("Read of packed %T did not equal original.\nExpected: %v\n     Got: %v", packet, packet, read)
		}
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
		if res := encodeLength(length); !bytes.Equal(res, encoded) {
			t.Errorf("encodeLength(%d) did not return [0x%X], but [0x%X]", length, encoded, res)
		}
	}
}

// isCopy checks if the original and copy are the same, recursively.
// It will log the field name and values if they are different.
// It will also log the field name and values if the pointer of the original and copy are the same.
func isCopy(t *testing.T, original, copy any, fieldName ...string) {
	t.Helper()

	log := func(field string, original, copy interface{}) {
		t.Logf("Field: %s", field)
		t.Logf("Original: %#v", original)
		t.Logf("Copy: %#v", copy)
	}

	originalValue := reflect.ValueOf(original)
	copyValue := reflect.ValueOf(copy)

	fullFieldName := ""
	if len(fieldName) > 0 {
		fullFieldName = fieldName[0]
		for _, name := range fieldName[1:] {
			fullFieldName += "." + name
		}
	}

	if originalValue.Kind() != copyValue.Kind() {
		log(fullFieldName, original, copy)
		t.Errorf("Kind of original and copy are different: %s != %s", originalValue.Kind(), copyValue.Kind())
	}

	switch originalValue.Kind() {
	case reflect.Ptr:
		if originalValue.Pointer() == copyValue.Pointer() {
			log(fullFieldName, original, copy)
			t.Errorf("Pointer of original and copy are the same: %x == %x", originalValue.Pointer(), copyValue.Pointer())
		}
		isCopy(t, originalValue.Elem().Interface(), copyValue.Elem().Interface(), append(fieldName, originalValue.Type().Elem().Name())...)
	case reflect.Slice:
		if originalValue.IsNil() && copyValue.IsNil() {
			return
		}
		if originalValue.IsNil() != copyValue.IsNil() {
			log(fullFieldName, original, copy)
			t.Errorf("IsNil of original and copy are different: %t != %t", originalValue.IsNil(), copyValue.IsNil())
		}
		if originalValue.Len() != copyValue.Len() {
			log(fullFieldName, original, copy)
			t.Errorf("Length of original and copy are different: %d != %d", originalValue.Len(), copyValue.Len())
		}
		if originalValue.Len() > 0 && originalValue.Pointer() == copyValue.Pointer() {
			log(fullFieldName, original, copy)
			t.Errorf("Pointer of original and copy are the same: %x == %x", originalValue.Pointer(), copyValue.Pointer())
		}
		for i := 0; i < originalValue.Len(); i++ {
			isCopy(t, originalValue.Index(i).Interface(), copyValue.Index(i).Interface(), append(fieldName, fmt.Sprintf("[%d]", i))...)
		}
	case reflect.Struct:
		for i := 0; i < originalValue.Type().NumField(); i++ {
			field := originalValue.Type().Field(i)
			isCopy(t, originalValue.Field(i).Interface(), copyValue.Field(i).Interface(), append(fieldName, field.Name)...)
		}
	default:
		if !reflect.DeepEqual(originalValue.Interface(), copyValue.Interface()) {
			log(fullFieldName, original, copy)
			t.Errorf("Values of original and copy are different: %v != %v", originalValue.Interface(), copyValue.Interface())
		}
	}
}

// fillMapsOrSlice creates valid pointer for map, slices or normal pointer if they are nil.
func fillMapsOrSlice(s any) {
	val := reflect.ValueOf(s).Elem()
	for i := range val.NumField() {
		field := val.Field(i)
		switch field.Kind() {
		case reflect.Ptr:
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
		case reflect.Slice:
			if field.IsNil() {
				field.Set(reflect.MakeSlice(field.Type(), 1, 1))
			}
		case reflect.Map:
			if field.IsNil() {
				field.Set(reflect.MakeMap(field.Type()))
			}
		case reflect.Struct:
			fillMapsOrSlice(field.Addr().Interface())
		}
	}
}

func TestPacketCopy(t *testing.T) {
	packets := []ControlPacket{
		NewControlPacket(Connack).(*ConnackPacket),
		NewControlPacket(Connect).(*ConnectPacket),
		NewControlPacket(Disconnect).(*DisconnectPacket),
		NewControlPacket(Pingreq).(*PingreqPacket),
		NewControlPacket(Pingresp).(*PingrespPacket),
		NewControlPacket(Puback).(*PubackPacket),
		NewControlPacket(Pubcomp).(*PubcompPacket),
		NewControlPacket(Publish).(*PublishPacket),
		NewControlPacket(Pubrec).(*PubrecPacket),
		NewControlPacket(Pubrel).(*PubrelPacket),
		NewControlPacket(Suback).(*SubackPacket),
		NewControlPacket(Subscribe).(*SubscribePacket),
		NewControlPacket(Unsuback).(*UnsubackPacket),
		NewControlPacket(Unsubscribe).(*UnsubscribePacket),
	}

	for _, packet := range packets {
		fillMapsOrSlice(packet)
		copy := packet.Copy()

		isCopy(t, packet, copy)
	}
}
