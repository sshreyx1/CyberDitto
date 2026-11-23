package twin

import "testing"

func TestCanTransition_HappyPathBuildToReady(t *testing.T) {
	path := []State{StateRequested, StateImaging, StateProvisioning, StateApplying, StateVerifying, StateReady}
	for i := 0; i < len(path)-1; i++ {
		if !CanTransition(path[i], path[i+1]) {
			t.Errorf("CanTransition(%s, %s) = false, want true", path[i], path[i+1])
		}
	}
}

func TestCanTransition_CannotSkipSteps(t *testing.T) {
	if CanTransition(StateRequested, StateReady) {
		t.Error("CanTransition(requested, ready) = true, want false (must not skip build steps)")
	}
	if CanTransition(StateImaging, StateRunning) {
		t.Error("CanTransition(imaging, running) = true, want false")
	}
}

func TestCanTransition_FailedIsReachableButNotReversible(t *testing.T) {
	for from := range transitions {
		if from == StateDestroyed || from == StateFailed {
			continue
		}
		if !CanTransition(from, StateFailed) {
			t.Errorf("CanTransition(%s, failed) = false, want true", from)
		}
	}
	for _, to := range []State{StateRequested, StateImaging, StateReady, StateRunning} {
		if CanTransition(StateFailed, to) {
			t.Errorf("CanTransition(failed, %s) = true, want false (no resuming a failed build)", to)
		}
	}
	if !CanTransition(StateFailed, StateDestroyed) {
		t.Error("CanTransition(failed, destroyed) = false, want true (a failed twin must still be cleanable up)")
	}
}

func TestCanTransition_DestroyedIsTerminal(t *testing.T) {
	if !StateDestroyed.Terminal() {
		t.Error("StateDestroyed.Terminal() = false, want true")
	}
	for to := range transitions {
		if CanTransition(StateDestroyed, to) {
			t.Errorf("CanTransition(destroyed, %s) = true, want false", to)
		}
	}
}

func TestCanTransition_StoppedCanRestartOrBeDestroyed(t *testing.T) {
	if !CanTransition(StateStopped, StateRunning) {
		t.Error("CanTransition(stopped, running) = false, want true")
	}
	if !CanTransition(StateStopped, StateDestroyed) {
		t.Error("CanTransition(stopped, destroyed) = false, want true")
	}
}

func TestState_Valid(t *testing.T) {
	if !StateReady.Valid() {
		t.Error("StateReady.Valid() = false, want true")
	}
	if State("made-up").Valid() {
		t.Error(`State("made-up").Valid() = true, want false`)
	}
}
