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
	"sync"
	"testing"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

func Test_newRouter(t *testing.T) {
	router := newRouter()
	if router == nil {
		t.Fatalf("router is nil")
	}
	if router.routes.Len() != 0 {
		t.Fatalf("router.routes was not empty")
	}
}

func Test_AddRoute(t *testing.T) {
	router := newRouter()
	cb := func(client Client, msg Message) {
	}
	router.addRoute("/alpha", cb)

	if router.routes.Len() != 1 {
		t.Fatalf("router.routes was wrong")
	}
}

func Test_AddRoute_Wildcards(t *testing.T) {
	router := newRouter()
	cb := func(client Client, msg Message) {
	}
	router.addRoute("#", cb)
	router.addRoute("topic1", cb)

	if router.routes.Len() != 2 {
		t.Fatalf("addRoute should only override routes on exact topic match")
	}
}

func Test_DeleteRoute_Wildcards(t *testing.T) {
	router := newRouter()
	cb := func(client Client, msg Message) {
	}
	router.addRoute("#", cb)
	router.addRoute("topic1", cb)
	router.deleteRoute("topic1")

	expected := "#"
	got := router.routes.Front().Value.(*route).topic
	if !(router.routes.Front().Value.(*route).topic == "#") {
		t.Fatalf("deleteRoute deleted wrong route when wildcards are used, got topic '%s', expected route with topic '%s'", got, expected)
	}
}

func Test_Match(t *testing.T) {
	router := newRouter()
	router.addRoute("/alpha", nil)

	if !router.routes.Front().Value.(*route).match("/alpha") {
		t.Fatalf("match function is bad")
	}

	if router.routes.Front().Value.(*route).match("alpha") {
		t.Fatalf("match function is bad")
	}
}

func Test_match(t *testing.T) {

	check := func(route, topic string, exp bool) {
		result := routeIncludesTopic(route, topic)
		if exp != result {
			t.Errorf("match was bad R: %v, T: %v, EXP: %v", route, topic, exp)
		}
	}

	// ** Basic **
	R := ""
	T := ""
	check(R, T, true)

	R = "x"
	T = ""
	check(R, T, false)

	R = ""
	T = "x"
	check(R, T, false)

	R = "x"
	T = "x"
	check(R, T, true)

	R = "x"
	T = "X"
	check(R, T, false)

	R = "alpha"
	T = "alpha"
	check(R, T, true)

	R = "alpha"
	T = "beta"
	check(R, T, false)

	// ** / **
	R = "/"
	T = "/"
	check(R, T, true)

	R = "/one"
	T = "/one"
	check(R, T, true)

	R = "/"
	T = "/two"
	check(R, T, false)

	R = "/two"
	T = "/"
	check(R, T, false)

	R = "/two"
	T = "two"
	check(R, T, false) // a leading "/" creates a different topic

	R = "/a/"
	T = "/a"
	check(R, T, false)

	R = "/a/"
	T = "/a/b"
	check(R, T, false)

	R = "/a/b"
	T = "/a/b"
	check(R, T, true)

	R = "/a/b/"
	T = "/a/b"
	check(R, T, false)

	R = "/a/b"
	T = "/R/b"
	check(R, T, false)

	// ** + **
	R = "/a/+/c"
	T = "/a/b/c"
	check(R, T, true)

	R = "/+/b/c"
	T = "/a/b/c"
	check(R, T, true)

	R = "/a/b/+"
	T = "/a/b/c"
	check(R, T, true)

	R = "/a/+/+"
	T = "/a/b/c"
	check(R, T, true)

	R = "/+/+/+"
	T = "/a/b/c"
	check(R, T, true)

	R = "/+/+/c"
	T = "/a/b/c"
	check(R, T, true)

	R = "/a/b/c/+" // different number of levels
	T = "/a/b/c"
	check(R, T, false)

	R = "+"
	T = "a"
	check(R, T, true)

	R = "/+"
	T = "a"
	check(R, T, false)

	R = "+/+"
	T = "/a"
	check(R, T, true)

	R = "+/+"
	T = "a"
	check(R, T, false)

	// ** # **
	R = "#"
	T = "/a/b/c"
	check(R, T, true)

	R = "/#"
	T = "/a/b/c"
	check(R, T, true)

	// R = "/#/" // not valid
	// T = "/a/b/c"
	// check(R, T, true)

	R = "/#"
	T = "/a/b/c"
	check(R, T, true)

	R = "/a/#"
	T = "/a/b/c"
	check(R, T, true)

	R = "/a/#"
	T = "/a/b/c"
	check(R, T, true)

	R = "/a/b/#"
	T = "/a/b/c"
	check(R, T, true)

	// ** unicode **
	R = "☃"
	T = "☃"
	check(R, T, true)

	R = "✈"
	T = "☃"
	check(R, T, false)

	R = "/☃/✈"
	T = "/☃/ッ"
	check(R, T, false)

	R = "#"
	T = "/☃/ッ"
	check(R, T, true)

	R = "/☃/+"
	T = "/☃/ッ/♫/ø/☹☹☹"
	check(R, T, false)

	R = "/☃/#"
	T = "/☃/ッ/♫/ø/☹☹☹"
	check(R, T, true)

	R = "/☃/ッ/♫/ø/+"
	T = "/☃/ッ/♫/ø/☹☹☹"
	check(R, T, true)

	R = "/☃/ッ/+/ø/☹☹☹"
	T = "/☃/ッ/♫/ø/☹☹☹"
	check(R, T, true)

	R = "/+/a/ッ/+/ø/☹☹☹"
	T = "/b/♫/ッ/♫/ø/☹☹☹"
	check(R, T, false)

	R = "/+/♫/ッ/+/ø/☹☹☹"
	T = "/b/♫/ッ/♫/ø/☹☹☹"
	check(R, T, true)
}

func Test_MatchAndDispatch(t *testing.T) {
	calledback := make(chan bool)

	cb := func(c Client, m Message) {
		calledback <- true
	}

	pub := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
	pub.Qos = 2
	pub.TopicName = "a"
	pub.Payload = []byte("foo")

	msgs := make(chan *packets.PublishPacket)

	router := newRouter()
	router.addRoute("a", cb)

	stopped := make(chan bool)
	go func() {
		router.matchAndDispatch(msgs, true, &client{oboundP: make(chan *PacketAndToken, 100)})
		stopped <- true
	}()
	msgs <- pub

	<-calledback

	close(msgs)

	select {
	case <-stopped:
		break
	case <-time.After(time.Second):
		t.Errorf("matchAndDispatch should have exited")
	}
}

func Test_SharedSubscription_MatchAndDispatch(t *testing.T) {
	calledback := make(chan bool)

	cb := func(c Client, m Message) {
		calledback <- true
	}

	pub := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
	pub.Qos = 2
	pub.TopicName = "a"
	pub.Payload = []byte("foo")

	msgs := make(chan *packets.PublishPacket)

	router := newRouter()
	router.addRoute("$share/az1/a", cb)

	stopped := make(chan bool)
	go func() {
		router.matchAndDispatch(msgs, true, &client{oboundP: make(chan *PacketAndToken, 100)})
		stopped <- true
	}()

	msgs <- pub

	<-calledback

	close(msgs)

	select {
	case <-stopped:
		break
	case <-time.After(time.Second):
		t.Errorf("matchAndDispatch should have exited")
	}

}

func Benchmark_MatchAndDispatch(b *testing.B) {
	calledback := make(chan bool, 1)

	cb := func(c Client, m Message) {
		calledback <- true
	}

	pub := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
	pub.TopicName = "a"
	pub.Payload = []byte("foo")

	msgs := make(chan *packets.PublishPacket, 1)

	router := newRouter()
	router.addRoute("a", cb)

	var wg sync.WaitGroup
	wg.Add(1)

	stopped := make(chan bool)
	go func() {
		wg.Done() // started
		<-router.matchAndDispatch(msgs, true, &client{oboundP: make(chan *PacketAndToken, 100)})
		stopped <- true
	}()

	wg.Wait()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msgs <- pub
		<-calledback
	}

	close(msgs)

	select {
	case <-stopped:
		break
	case <-time.After(time.Second):
		b.Errorf("matchAndDispatch should have exited")
	}
}
