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

package mqtt

import (
	"errors"
	"testing"
	"time"
)

func TestWaitTimeout(t *testing.T) {
	b := baseToken{}

	if b.WaitTimeout(time.Second) {
		t.Fatal("Should have failed")
	}

	// Now lets confirm that WaitTimeout returns
	// setError() grabs the mutex which previously caused issues
	// when there is a result (it returns true in this case)
	b = baseToken{complete: make(chan struct{})}
	go func(bt *baseToken) {
		bt.setError(errors.New("test error"))
	}(&b)
	if !b.WaitTimeout(5 * time.Second) {
		t.Fatal("Should have succeeded")
	}
}

func TestWaitTokenTimeout(t *testing.T) {
	b := baseToken{}

	if !errors.Is(WaitTokenTimeout(&b, time.Second), TimedOut) {
		t.Fatal("Should have failed")
	}

	// Now let's confirm that WaitTimeout returns correct error
	b = baseToken{complete: make(chan struct{})}
	testError := errors.New("test")
	go func(bt *baseToken) {
		bt.setError(testError)
	}(&b)
	if !errors.Is(WaitTokenTimeout(&b, 5*time.Second), testError) {
		t.Fatal("Unexpected error received")
	}
}
