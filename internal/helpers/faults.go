package helpers

import (
	"log"
	"sync"
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
	NextWaiter(point string)
	JustRun()
	AllowAll(point string)
	AllocatedWaiter(point string) PairedWaiter
	AllocatedWaiterOrNil(point string, waitFor time.Duration) PairedWaiter
}

type TestingFaultInjection struct {
	t           Fatals
	exclusion   sync.Mutex
	allocations map[string]chan PairedWaiter
	letRun      bool
	allowing    map[string]bool
}

// JustRun implements FaultInjection.
func (finj *TestingFaultInjection) JustRun() {
	finj.letRun = true
	log.Printf("releasing all to run freely")
}

func (finj *TestingFaultInjection) AllowAll(point string) {
	finj.allowing[point] = true
}

// AllocatedWaiter implements FaultInjection.
func (finj *TestingFaultInjection) AllocatedWaiter(point string) PairedWaiter {
	r := finj.AllocatedWaiterOrNil(point, 5*time.Second)
	if r == nil {
		finj.t.Fatalf("waiter %s had not been allocated after 5s", point)
	}
	return r
}

// AllocatedWaiter implements FaultInjection.
func (finj *TestingFaultInjection) AllocatedWaiterOrNil(point string, waitFor time.Duration) PairedWaiter {
	finj.ensure(point)
	select {
	case <-time.After(waitFor):
		log.Printf("non-allocated(%s) after %d", point, waitFor)
		return nil
	case ret := <-finj.allocations[point]:
		log.Printf("allocated(%s) was %p", point, ret)
		return ret
	}
}

// NextWaiter implements FaultInjection.
func (finj *TestingFaultInjection) NextWaiter(point string) {
	if finj.letRun || finj.allowing[point] {
		log.Printf("running through allocation for %s", point)
		return
	}
	ret := &SimplePairedWaiter{t: finj.t, notifyMe: make(chan struct{}), delay: 10 * time.Second}
	finj.ensure(point)
	finj.allocations[point] <- ret
	log.Printf("next(%s) allocated %p, waiting ...", point, ret)
	ret.Wait()
	log.Printf("released(%s, %p)", point, ret)
}

func (finj *TestingFaultInjection) ensure(point string) {
	finj.exclusion.Lock()
	defer finj.exclusion.Unlock()
	if finj.allocations[point] == nil {
		finj.allocations[point] = make(chan PairedWaiter)
	}
}

func FaultInjectionLibrary(t Fatals) FaultInjection {
	return &TestingFaultInjection{t: t, allocations: make(map[string]chan PairedWaiter, 10), allowing: make(map[string]bool)}
}

type InactiveFaultInjection struct{}

// AllowAll implements FaultInjection.
func (i *InactiveFaultInjection) AllowAll(point string) {
	panic("this should only be called from tests")
}

// JustRun implements FaultInjection.
func (i *InactiveFaultInjection) JustRun() {
	panic("this should only be called from tests")
}

// AllocatedWaiter implements FaultInjection.
func (i *InactiveFaultInjection) AllocatedWaiter(point string) PairedWaiter {
	panic("this should only be called from test methods, I think")
}

// AllocatedWaiterOrNil implements FaultInjection.
func (i *InactiveFaultInjection) AllocatedWaiterOrNil(point string, waitFor time.Duration) PairedWaiter {
	panic("this should only be called from test methods, I think")
}

// NextWaiter implements FaultInjection.
func (i *InactiveFaultInjection) NextWaiter(point string) {
}

func IgnoreFaultInjection() FaultInjection {
	return &InactiveFaultInjection{}
}
