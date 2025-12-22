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

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/url"

	"golang.org/x/net/proxy"

	// "log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	MQTT "github.com/liooooo29/paho.mqtt.golang"
)

func onMessageReceived(_ MQTT.Client, message MQTT.Message) {
	fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
}

func init() {
	// Pre-register custom HTTP proxy dialers for use with proxy.FromEnvironment
	proxy.RegisterDialerType("http", newHTTPProxy)
	proxy.RegisterDialerType("https", newHTTPProxy)
}

/**
 * Illustrates how to make an MQTT connection with HTTP proxy CONNECT support.
 * Specify proxy via environment variable: eg: ALL_PROXY=https://proxy_host:port
 */
func main() {
	MQTT.DEBUG = log.New(os.Stdout, "", 0)
	MQTT.ERROR = log.New(os.Stderr, "", 0)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	hostname, _ := os.Hostname()

	server := flag.String("server", "tcp://127.0.0.1:1883", "The full URL of the MQTT server to "+
		"connect to ex: tcp://127.0.0.1:1883")
	topic := flag.String("topic", "#", "Topic to subscribe to")
	qos := flag.Int("qos", 0, "The QoS to subscribe to messages at")
	clientid := flag.String("clientid", hostname+strconv.Itoa(time.Now().Second()), "A clientid for the connection")
	username := flag.String("username", "", "A username to authenticate to the MQTT server")
	password := flag.String("password", "", "Password to match username")
	token := flag.String("token", "", "An optional token credential to authenticate with")
	skipVerify := flag.Bool("skipVerify", false, "Controls whether TLS certificate is verified")
	flag.Parse()

	connOpts := MQTT.NewClientOptions().AddBroker(*server).
		SetClientID(*clientid).
		SetCleanSession(true).
		SetProtocolVersion(4)

	if *username != "" {
		connOpts.SetUsername(*username)
		if *password != "" {
			connOpts.SetPassword(*password)
		}
	} else if *token != "" {
		connOpts.SetCredentialsProvider(func() (string, string) {
			return "unused", *token
		})
	}

	connOpts.SetTLSConfig(&tls.Config{InsecureSkipVerify: *skipVerify, ClientAuth: tls.NoClientCert})

	connOpts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(*topic, byte(*qos), onMessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	// Illustrates customized TLS configuration prior to connection attempt
	connOpts.OnConnectAttempt = func(broker *url.URL, tlsCfg *tls.Config) *tls.Config {
		cfg := tlsCfg.Clone()
		cfg.ServerName = broker.Hostname()
		return cfg
	}

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Printf("Connected to %s\n", *server)
	}

	<-c
}
