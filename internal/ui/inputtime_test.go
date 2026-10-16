// Copyright 2026 The Ebitengine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ui

import "testing"

// Under RunGameEmbedded the host's own event pump (CEF's message loop) can
// deliver a key while an Update runs on the game goroutine. Such a press was
// stamped with the running tick, whose input state was already read, so no
// Update ever saw it as just pressed: a Space press sometimes did nothing.

func TestInputTime_AnEventDuringAnUpdateBelongsToTheNextTick(t *testing.T) {
	u := &UserInterface{}
	var state InputState

	// Tick 0: the Update read its input, then the key goes down while it runs.
	u.beginTickInput()
	state.setKeyPressed(KeySpace, u.InputTime())
	if state.IsKeyJustPressed(KeySpace, 0) {
		t.Fatal("the running tick's input was already read; it cannot see this press")
	}
	u.incrementTick()
	if !state.IsKeyJustPressed(KeySpace, 1) {
		t.Fatal("a press that came during an Update is never just pressed")
	}
}

func TestInputTime_ATapDuringAnUpdateIsStillJustPressedNextTick(t *testing.T) {
	u := &UserInterface{}
	var state InputState

	u.beginTickInput()
	state.setKeyPressed(KeySpace, u.InputTime())
	state.setKeyReleased(KeySpace, u.InputTime())
	u.incrementTick()
	if !state.IsKeyJustPressed(KeySpace, 1) {
		t.Fatal("a quick tap during an Update is lost")
	}
	if state.IsKeyPressed(KeySpace, 2) {
		t.Fatal("the tapped key is still held a tick later")
	}
}

func TestInputTime_AnEventBetweenUpdatesBelongsToTheComingTick(t *testing.T) {
	u := &UserInterface{}
	var state InputState

	u.beginTickInput()
	u.incrementTick() // tick 0 done; tick 1 has not read its input yet
	state.setKeyPressed(KeySpace, u.InputTime())
	if !state.IsKeyJustPressed(KeySpace, 1) {
		t.Fatal("a press between two Updates is not just pressed in the next one")
	}
}

func TestInputTime_StampsNeverRepeatAcrossTheTickEnd(t *testing.T) {
	u := &UserInterface{}
	var state InputState

	// Pressed during the Update of tick 0, released right after the tick ended:
	// the release must stamp later than the press, or it is dropped as a
	// duplicate release (#3326) and the key stays held for ever.
	u.beginTickInput()
	pressed := u.InputTime()
	state.setKeyPressed(KeySpace, pressed)
	u.incrementTick()
	released := u.InputTime()
	if released <= pressed {
		t.Fatalf("release stamped %d, not after the press %d", released, pressed)
	}
	state.setKeyReleased(KeySpace, released)
	if state.IsKeyPressed(KeySpace, 2) {
		t.Fatal("the release was dropped: the key stays held")
	}
}
