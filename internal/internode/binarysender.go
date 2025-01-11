package internode

type BinarySender interface {
	Send(path string, blob []byte)
}
