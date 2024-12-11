package harness

type Config interface {
}

type Client interface {
	Begin()
	WaitFor()
}

func ReadConfig() *Config {
	return nil
}

func StartNodes(c *Config) {

}

func PrepareClients(c *Config) []Client {
	return make([]Client, 0)
}
