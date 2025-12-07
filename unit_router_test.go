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
	router := newRouter(noopSLogger)
	if router == nil {
		t.Fatalf("router is nil")
	}
	if router.routes.Len() != 0 {
		t.Fatalf("router.routes was not empty")
	}
}

func Test_AddRoute(t *testing.T) {
	router := newRouter(noopSLogger)
	cb := func(client Client, msg Message) {
	}
	router.addRoute("/alpha", cb)

	if router.routes.Len() != 1 {
		t.Fatalf("router.routes was wrong")
	}
}

func Test_AddRoute_Wildcards(t *testing.T) {
	router := newRouter(noopSLogger)
	cb := func(client Client, msg Message) {
	}
	router.addRoute("#", cb)
	router.addRoute("topic1", cb)

	if router.routes.Len() != 2 {
		t.Fatalf("addRoute should only override routes on exact topic match")
	}
}

func Test_DeleteRoute_Wildcards(t *testing.T) {
	router := newRouter(noopSLogger)
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
	router := newRouter(noopSLogger)
	router.addRoute("/alpha", nil)

	if !router.routes.Front().Value.(*route).match("/alpha") {
		t.Fatalf("match function is bad")
	}

	if router.routes.Front().Value.(*route).match("alpha") {
		t.Fatalf("match function is bad")
	}
}

func Test_matchWithWildcardTopicFilter(t *testing.T) {
	cases := []struct {
		name     string
		filter   string
		topic    string
		expected bool
	}{
		// OK cases with wildcards
		{"OK_WithPlus", "example/+/temp/+", "example/gauge/temp/25", true},
		{"OK_WithPlus_$", "$sys/+/temp/+", "$sys/gauge/temp/25", true},
		{"OK_WithSharp", "example/machines/#", "example/machines/robot-2/voltage", true},
		{"OK_WithSharp_$", "$SYS/machines/#", "$SYS/machines/robot-2/voltage", true},

		// OK cases with filter starting with wildcards
		{"OK_OnlyPlus", "+", "sensor", true},
		{"OK_StartWithPlus", "+/example", "sensor/example", true},
		{"OK_OnlySharp", "#", "sensor/very/long/topic", true},
		{"OK_SharedTopicWithPlus", "$share/group/+/example/+/data", "sensor/example/something/data", true},
		{"OK_SharedTopicWithSharp", "$share/group/#", "sensor/this/is/shared", true},

		// NG cases with filter starting with wildcards and topics starting with $
		{"NG_OnlyPlus", "+", "$SYS", false},
		{"NG_StartWithPlus", "+/example/sensor/+", "$VENDOR/example/sensor/data", false},
		{"NG_OnlySharp", "#", "$sys/very/long/topic", false},
		{"NG_SharedTopicWithSharp", "$share/group/#", "$vendor/this/is/shared", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			router := newRouter(noopSLogger)
			router.addRoute(c.filter, nil)
			result := router.routes.Front().Value.(*route).match(c.topic)
			if result != c.expected {
				t.Errorf("match function returned %v, expected %v for filter '%s' and topic '%s'", result, c.expected, c.filter, c.topic)
			}
		})
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

	router := newRouter(noopSLogger)
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

	router := newRouter(noopSLogger)
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

	router := newRouter(noopSLogger)
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
