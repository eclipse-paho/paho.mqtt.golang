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
	"testing"
)

func Test_getID(t *testing.T) {
	mids := &messageIds{index: make(map[uint16]tokenCompletor), logger: noopSLogger}

	i1 := mids.getID(&DummyToken{})

	if i1 != 1 {
		t.Fatalf("i1 was wrong: %v", i1)
	}

	i2 := mids.getID(&DummyToken{})

	if i2 != 2 {
		t.Fatalf("i2 was wrong: %v", i2)
	}

	for i := uint16(3); i < 100; i++ {
		id := mids.getID(&DummyToken{})
		if id != i {
			t.Fatalf("id was wrong expected %v got %v", i, id)
		}
	}
}

func Test_freeID(t *testing.T) {
	mids := &messageIds{index: make(map[uint16]tokenCompletor), logger: noopSLogger}

	i1 := mids.getID(&DummyToken{})
	mids.freeID(i1)

	if i1 != 1 {
		t.Fatalf("i1 was wrong: %v", i1)
	}

	// The below may be needed for a specific test but leaving it in permanently makes output confusing
	// i2 := mids.getID(&DummyToken{})
	// fmt.Printf("i2: %v\n", i2)
}

func Test_noFreeID(t *testing.T) {
	var d DummyToken

	mids := &messageIds{index: make(map[uint16]tokenCompletor), logger: noopSLogger}

	for i := midMin; i != 0; i++ {
		// Uncomment to see all message IDS log.Println(i)
		mids.index[i] = &d
	}

	mid := mids.getID(&d)
	if mid != 0 {
		t.Errorf("shouldn't be any mids left")
	}

	mid = mids.getID(&d)
	if mid != 0 {
		t.Errorf("shouldn't be any mids left")
	}
}
