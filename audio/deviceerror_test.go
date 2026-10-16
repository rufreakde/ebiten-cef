// Copyright 2026 The Ebiten Authors
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

package audio_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

var errNoDevice = errors.New("no sound device")

func TestDeviceErrorEndsTheGameWithoutAHandler(t *testing.T) {
	audio.ResetHooksForTesting()
	setup()
	defer teardown()
	defer audio.SetDeviceErrForTesting(nil)

	if err := audio.UpdateForTesting(); err != nil {
		t.Fatal(err)
	}
	audio.SetDeviceErrForTesting(errNoDevice)
	if err := audio.UpdateForTesting(); !errors.Is(err, errNoDevice) {
		t.Errorf("update error = %v, want the device error", err)
	}
}

func TestDeviceErrorHandlerKeepsTheGameRunningSilently(t *testing.T) {
	audio.ResetHooksForTesting()
	setup()
	defer teardown()
	defer audio.SetDeviceErrForTesting(nil)

	var got []error
	context.SetDeviceErrorHandler(func(err error) { got = append(got, err) })
	if err := audio.UpdateForTesting(); err != nil {
		t.Fatal(err)
	}
	audio.SetDeviceErrForTesting(errNoDevice)
	for range 3 {
		if err := audio.UpdateForTesting(); err != nil {
			t.Fatalf("update error = %v, want nil once a handler is set", err)
		}
	}
	if len(got) != 1 || !errors.Is(got[0], errNoDevice) {
		t.Errorf("handler got %v, want the device error once", got)
	}

	// Players still work against a silenced context.
	p, err := context.NewPlayer(bytes.NewReader(make([]byte, 4)))
	if err != nil {
		t.Fatal(err)
	}
	p.Play()
	if err := audio.UpdateForTesting(); err != nil {
		t.Errorf("update after Play = %v", err)
	}
	_ = p.Close()
}
