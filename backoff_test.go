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
 *    Matt Brittan
 *    Daichi Tomaru
 */

package mqtt

import (
	"testing"
	"time"
)

func TestGetBackoffSleepTime(t *testing.T) {
	// Test for adding new situation
	controller := newBackoffController()
	if s, c := controller.getBackoffSleepTime("not-exist", 1 * time.Second, 5 * time.Second, 1 * time.Second, false); !((s == 1 * time.Second) && !c) {
		t.Errorf("When new situation is added, period should be initSleepPeriod and naturally it shouldn't be continual error. s:%d c%t", s, c)
	}

	// Test for the continual error in the same situation and suppression of sleep period by maxSleepPeriod
	controller.getBackoffSleepTime("multi", 10 * time.Second, 30 * time.Second, 1 * time.Second, false)
	if s, c := controller.getBackoffSleepTime("multi", 10 * time.Second, 30 * time.Second, 1 * time.Second, false); !((s == 20 * time.Second) && c) {
		t.Errorf("When same situation is called again, period should be increased and it should be regarded as a continual error. s:%d c%t", s, c)
	}
	if s, c := controller.getBackoffSleepTime("multi", 10 * time.Second, 30 * time.Second, 1 * time.Second, false); !((s == 30 * time.Second) && c) {
		t.Errorf("A same situation is called three times. 10 * 2 * 2 = 40 but maxSleepPeriod is 30. So the next period should be 30. s:%d c%t", s, c)
	}

	// Test for initialization by elapsed time.
	controller.getBackoffSleepTime("elapsed", 1 * time.Second, 128 * time.Second, 1 * time.Second, false)
	controller.getBackoffSleepTime("elapsed", 1 * time.Second, 128 * time.Second, 1 * time.Second, false)
	time.Sleep((1 * 2 + 1 * 2 + 1) * time.Second)
	if s, c := controller.getBackoffSleepTime("elapsed", 1 * time.Second, 128 * time.Second, 1 * time.Second, false); !((s == 1 * time.Second) && !c) {
		t.Errorf("Initialization should be triggered by elapsed time. s:%d c%t", s, c)
	}

	// Test when initial and max period is same.
	controller.getBackoffSleepTime("same", 2 * time.Second, 2 * time.Second, 1 * time.Second, false)
	if s, c := controller.getBackoffSleepTime("same", 2 * time.Second, 2 * time.Second, 1 * time.Second, false); !((s == 2 * time.Second) && c) {
		t.Errorf("Sleep time should be always 2. s:%d c%t", s, c)
	}

	// Test when initial period > max period.
	controller.getBackoffSleepTime("bigger", 5 * time.Second, 2 * time.Second, 1 * time.Second, false)
	if s, c := controller.getBackoffSleepTime("bigger", 5 * time.Second, 2 * time.Second, 1 * time.Second, false); !((s == 2 * time.Second) && c) {
		t.Errorf("Sleep time should be 2. s:%d c%t", s, c)
	}

	// Test when first sleep is skipped.
	if s, c := controller.getBackoffSleepTime("skip", 3 * time.Second, 12 * time.Second, 1 * time.Second, true); !((s == 0) && !c) {
		t.Errorf("Sleep time should be 0 because of skip. s:%d c%t", s, c)
	}
	if s, c := controller.getBackoffSleepTime("skip", 3 * time.Second, 12 * time.Second, 1 * time.Second, true); !((s == 3 * time.Second) && c) {
		t.Errorf("Sleep time should be 3. s:%d c%t", s, c)
	}
}
