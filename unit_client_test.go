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
	"log"
	"net/http"
	_ "net/http/pprof"
	"testing"
	"time"
)

func init() {
	// Logging is off by default as this makes things simpler when you just want to confirm that tests pass
	// DEBUG = log.New(os.Stderr, "DEBUG    ", log.Ltime)
	// WARN = log.New(os.Stderr, "WARNING  ", log.Ltime)
	// CRITICAL = log.New(os.Stderr, "CRITICAL ", log.Ltime)
	// ERROR = log.New(os.Stderr, "ERROR    ", log.Ltime)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func Test_NewClient_simple(t *testing.T) {
	ops := NewClientOptions().SetClientID("foo").AddBroker("tcp://10.10.0.1:1883")
	c := NewClient(ops).(*client)

	if c == nil {
		t.Fatalf("ops is nil")
	}

	if c.options.ClientID != "foo" {
		t.Fatalf("bad client id")
	}

	if c.options.Servers[0].Scheme != "tcp" {
		t.Fatalf("bad server scheme")
	}

	if c.options.Servers[0].Host != "10.10.0.1:1883" {
		t.Fatalf("bad server host")
	}
}

func Test_NewClient_optionsReader(t *testing.T) {
	ops := NewClientOptions().SetClientID("foo").AddBroker("tcp://10.10.0.1:1883").SetMaxIncomingPacketSize(4096)
	c := NewClient(ops).(*client)

	if c == nil {
		t.Fatalf("ops is nil")
	}

	rOps := c.OptionsReader()
	cid := rOps.ClientID()

	if cid != "foo" {
		t.Fatalf("unable to read client ID")
	}

	servers := rOps.Servers()
	broker := servers[0]
	if broker.Hostname() != "10.10.0.1" {
		t.Fatalf("unable to read hostname")
	}
	if rOps.MaxIncomingPacketSize() != 4096 {
		t.Fatalf("unable to read maximum incoming packet size")
	}

}

func Test_isConnection(t *testing.T) {
	ops := NewClientOptions()
	c := NewClient(ops)

	c.(*client).status.forceConnectionStatus(connected)
	if !c.IsConnectionOpen() {
		t.Fail()
	}
}

func Test_isConnectionOpenNegative(t *testing.T) {
	ops := NewClientOptions()
	c := NewClient(ops)

	c.(*client).status.forceConnectionStatus(reconnecting)
	if c.IsConnectionOpen() {
		t.Fail()
	}
	c.(*client).status.forceConnectionStatus(connecting)
	if c.IsConnectionOpen() {
		t.Fail()
	}
	c.(*client).status.forceConnectionStatus(disconnected)
	if c.IsConnectionOpen() {
		t.Fail()
	}
}

// Test_PublishQoS0NotConnected checks that the token returned when publishing a QoS 0 message, whilst the connection
// is not up but IsConnected() returns true, completes (QoS 0 messages are not stored, so nothing else would
// complete the token). See issue #798.
func Test_PublishQoS0NotConnected(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*ClientOptions, *client)
	}{
		{
			name: "connecting with ConnectRetry",
			setup: func(o *ClientOptions, c *client) {
				o.SetConnectRetry(true)
				c.status.forceConnectionStatus(connecting)
			},
		},
		{
			name: "disconnecting with reconnect planned",
			setup: func(o *ClientOptions, c *client) {
				o.SetAutoReconnect(true)
				c.status.forceConnectionStatus(connected)
				if _, err := c.status.ConnectionLost(true); err != nil {
					t.Fatalf("ConnectionLost returned unexpected error: %v", err)
				}
			},
		},
		{
			name: "reconnecting",
			setup: func(o *ClientOptions, c *client) {
				o.SetAutoReconnect(true)
				c.status.forceConnectionStatus(reconnecting)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ops := NewClientOptions()
			ops.SetWriteTimeout(time.Second) // Avoid a long wait should the message be passed to obound
			c := NewClient(ops).(*client)
			tt.setup(&c.options, c)
			if !c.IsConnected() {
				t.Fatalf("expected IsConnected() to return true")
			}

			token := c.Publish("test/topic", 0, false, "payload")
			if !token.WaitTimeout(5 * time.Second) {
				t.Fatalf("QoS 0 publish token did not complete")
			}
			if token.Error() != nil {
				t.Fatalf("unexpected error: %v", token.Error())
			}
		})
	}
}
