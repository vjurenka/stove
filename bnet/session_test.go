package bnet

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTransitions(t *testing.T) {
	runtime.GOMAXPROCS(8)

	s := &Session{}
	s.stateChange = sync.NewCond(&s.stateMutex)
	s.stateListeners = map[int]int32{}
	s.state = StateConnecting

	numWaits := int32(1000)
	j := int32(0)
	for i := int32(0); i < numWaits; i++ {
		go func() {
			s.WaitForTransition(StateDisconnected)
			atomic.AddInt32(&j, 1)
		}()
	}
	for s.stateListeners[StateDisconnected] < numWaits {
		time.Sleep(time.Millisecond)
	}
	// each goroutine needs to get its PC through the cond.Wait() call, so we
	// grab a momentary lock on the mutex to ensure everything gets through.
	s.stateMutex.Lock()
	done := make(chan struct{})
	go func() {
		s.Transition(StateDisconnected)
		s.Transition(StateConnecting)
		done <- struct{}{}
	}()
	s.stateMutex.Unlock()
	<-done
	if j != numWaits {
		t.Errorf("failed: %d != %d", j, numWaits)
	}
}
