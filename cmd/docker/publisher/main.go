/*
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * and Eclipse Distribution License v1.0 which accompany this distribution.
 *
 * The Eclipse Public License is available at
 *    https://www.eclipse.org/legal/epl-2.0/
 * and the Eclipse Distribution License is available at
 *   http://www.eclipse.org/org/documents/edl-v10.php.

 * Contributors:
 *    Matt Brittan
 */

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/liooooo29/paho.mqtt.golang"
)

// Connect to the broker and publish a message periodically

const (
	TOPIC         = "topic1"
	QOS           = 1
	SERVERADDRESS = "tcp://mosquitto:1883"
	DELAY         = time.Second
	CLIENTID      = "mqtt_publisher"

	WRITETOLOG = true // If true then published messages will be written to the console
)

func main() {
	// Enable logging by uncommenting the below
	// mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	// mqtt.CRITICAL = log.New(os.Stdout, "[CRITICAL] ", 0)
	// mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(SERVERADDRESS)
	opts.SetClientID(CLIENTID)

	opts.SetOrderMatters(false)       // Allow out of order messages (use this option unless in order delivery is essential)
	opts.ConnectTimeout = time.Second // Minimal delays on connect
	opts.WriteTimeout = time.Second   // Minimal delays on writes
	opts.KeepAlive = 10               // Keepalive every 10 seconds so we quickly detect network outages
	opts.PingTimeout = time.Second    // local broker so response should be quick

	// Automate connection management (will keep trying to connect and will reconnect if network drops)
	opts.ConnectRetry = true
	opts.AutoReconnect = true

	// Log events
	opts.OnConnectionLost = func(cl mqtt.Client, err error) {
		fmt.Println("connection lost")
	}
	opts.OnConnect = func(mqtt.Client) {
		fmt.Println("connection established")
	}
	opts.OnReconnecting = func(mqtt.Client, *mqtt.ClientOptions) {
		fmt.Println("attempting to reconnect")
	}

	//
	// Connect to the broker
	//
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("Connection is up")

	//
	// Publish messages until we receive a signal
	//
	done := make(chan struct{})
	var wg sync.WaitGroup

	// The message could be anything; lets make it JSON containing a simple count (makes it simpler to track the messages)
	type msg struct {
		Count uint64
	}

	wg.Add(1)
	go func() {
		var count uint64
		for {
			select {
			case <-time.After(DELAY):
				count += 1
				msg, err := json.Marshal(msg{Count: count})
				if err != nil {
					panic(err)
				}

				if WRITETOLOG {
					fmt.Printf("sending message: %s\n", msg)
				}
				t := client.Publish(TOPIC, QOS, false, msg)
				// Handle the token in a go routine so this loop keeps sending messages regardless of delivery status
				go func() {
					_ = t.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
					if t.Error() != nil {
						fmt.Printf("ERROR PUBLISHING: %s\n", err)
					}
				}()
			case <-done:
				fmt.Println("publisher done")
				wg.Done()
				return
			}
		}
	}()

	// Wait for a signal before exiting
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	fmt.Println("signal caught - exiting")

	close(done)
	wg.Wait()
	fmt.Println("shutdown complete")
}
