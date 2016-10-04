package goremproc

type RemoteIPConn interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}

type RemoteIPDialer struct {
	cc *ControlChannel
}

type RemoteIPDialRequest struct {
	network string
	address string
}

type RemoteIPDialResult struct {
}

func (d *RemoteIPDialer) Dial(network, address string) (*RemoteIPConn, error) {
	//d.cc.issue(r)
	return nil, nil
}

func (c *ControlChannel) NewRemoteIPDialer() (d *RemoteIPDialer, err error) {
	d.cc = c
	return
}
