package remote

type Client interface {
	Run() error
	Close()
	SendMsg(string, []byte) error
}
