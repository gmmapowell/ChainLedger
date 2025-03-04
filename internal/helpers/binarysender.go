package helpers

type BinarySender interface {
	Send(path string, blob []byte)
}
