package types

// The syntax for sending "signal" messages through a channel is ugly,
// so I'm going to fix it here

type Signal chan struct{}
type SendSignal chan<- struct{}
type OnSignal <-chan struct{}

type PingBack chan SendSignal

func (s Signal) Sender() SendSignal {
	var k chan struct{} = s
	return SendSignal(k)
}

func (s Signal) Reader() OnSignal {
	var k chan struct{} = s
	return OnSignal(k)
}

func (s SendSignal) Send() {
	s <- struct{}{}
}
