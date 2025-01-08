package helpers

import (
	"time"
)

type PairedWaiter interface {
	Wait()
	Release()
}

type SimplePairedWaiter struct {
	t        Fatals
	notifyMe chan struct{}
	delay    time.Duration
}

func (spw *SimplePairedWaiter) Wait() {
	select {
	case <-time.After(spw.delay):
		spw.t.Fatalf("waited for %d but not notified", spw.delay)
	case <-spw.notifyMe:
	}
}

func (spw SimplePairedWaiter) Release() {
	spw.notifyMe <- struct{}{}
}

type FaultInjection interface {
	NextWaiter()
	JustRun()
	AllocatedWaiter() PairedWaiter
	AllocatedWaiterOrNil(waitFor time.Duration) PairedWaiter
}

type TestingFaultInjection struct {
	t           Fatals
	allocations chan PairedWaiter
	letRun      bool
}

// JustRun implements FaultInjection.
func (t *TestingFaultInjection) JustRun() {
	t.letRun = true
}

// AllocatedWaiter implements FaultInjection.
func (t *TestingFaultInjection) AllocatedWaiter() PairedWaiter {
	r := t.AllocatedWaiterOrNil(5 * time.Second)
	if r == nil {
		t.t.Fatalf("waiter had not been allocated after 5s")
	}
	return r
}

// AllocatedWaiter implements FaultInjection.
func (t *TestingFaultInjection) AllocatedWaiterOrNil(waitFor time.Duration) PairedWaiter {
	select {
	case <-time.After(waitFor):
		return nil
	case ret := <-t.allocations:
		return ret
	}
}

// NextWaiter implements FaultInjection.
func (t *TestingFaultInjection) NextWaiter() {
	if t.letRun {
		return
	}
	ret := &SimplePairedWaiter{t: t.t, notifyMe: make(chan struct{}), delay: 10 * time.Second}
	t.allocations <- ret
	ret.Wait()
}

func FaultInjectionLibrary(t Fatals) FaultInjection {
	return &TestingFaultInjection{t: t, allocations: make(chan PairedWaiter, 10)}
}

type InactiveFaultInjection struct{}

// JustRun implements FaultInjection.
func (i *InactiveFaultInjection) JustRun() {
	panic("this should only be called from tests")
}

// AllocatedWaiter implements FaultInjection.
func (i *InactiveFaultInjection) AllocatedWaiter() PairedWaiter {
	panic("this should only be called from test methods, I think")
}

// AllocatedWaiterOrNil implements FaultInjection.
func (i *InactiveFaultInjection) AllocatedWaiterOrNil(waitFor time.Duration) PairedWaiter {
	panic("this should only be called from test methods, I think")
}

// NextWaiter implements FaultInjection.
func (i *InactiveFaultInjection) NextWaiter() {
}

func IgnoreFaultInjection() FaultInjection {
	return &InactiveFaultInjection{}
}
