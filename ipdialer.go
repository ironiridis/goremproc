package goremproc

import (
	"encoding/json"
	"errors"
)

// A RemoteIPConn implements simple ReadWriteCloser semantics.
type RemoteIPConn struct {
	handle uint64
	cc     *ControlChannel
}

// Read reads from the socket opened on the remote endpoint.
func (c *RemoteIPConn) Read(b []byte) (int, error) {
	return 0, ErrorStub
}

// Write writes to the socket opened on the remote endpoint.
func (c *RemoteIPConn) Write(b []byte) (int, error) {
	return 0, ErrorStub
}

// Close closes the socket on the remote endpoint. The RemoteIPConn will be
// invalid afterward.
func (c *RemoteIPConn) Close() error {
	return ErrorStub
}

// A RemoteIPDialer allows you to open an IP connection from a remote endpoint.
type RemoteIPDialer struct {
	cc *ControlChannel
}

// Dial attemps to open a socket from the perspective of a remote endpoint.
func (d *RemoteIPDialer) Dial(network, address string) (*RemoteIPConn, error) {
	ch, err := d.cc.issue(&remoteIPDialRequest{Network: network, Address: address})
	if err != nil {
		return nil, err
	}

	chres, ok := <-ch
	if !ok {
		return nil, ErrorRequestCancelled
	}
	res := chres.(*remoteIPDialResult)
	if !res.Success {
		return nil, errors.New(res.Error)
	}
	return &RemoteIPConn{handle: res.Handle}, nil
}

// NewRemoteIPDialer creates a reusable interface for Dialing new connections.
func (c *ControlChannel) NewRemoteIPDialer() (d *RemoteIPDialer, err error) {
	d.cc = c
	return
}

type remoteIPDialRequest struct {
	Network string
	Address string
}

func (r *remoteIPDialRequest) t() string { return "remoteIPDialRequest" }
func (r *remoteIPDialRequest) MarshalText() ([]byte, error) {
	return json.Marshal(r)
}
func (r *remoteIPDialRequest) UnmarshalText(text []byte) error {
	return json.Unmarshal(text, r)
}

type remoteIPDialResult struct {
	Success bool
	Error   string `json:",omitempty"`
	Handle  uint64 `json:",omitempty"`
}

func (r *remoteIPDialResult) MarshalText() ([]byte, error) {
	return json.Marshal(r)
}
func (r *remoteIPDialResult) UnmarshalText(text []byte) error {
	return json.Unmarshal(text, r)
}
