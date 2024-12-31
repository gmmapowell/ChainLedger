package helpers

type Fatals interface {
	Fatalf(format string, args ...any)
	Log(args ...any)
	Logf(format string, args ...any)
	Fail()
}
