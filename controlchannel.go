package goremproc

import "sync"
import "errors"
import "encoding"

// ErrorControlChannelNotOpen is returned for any API call that cannot
// continue because either the ControlChannel is not ready or has encountered
// an error.
var ErrorControlChannelNotOpen = errors.New("Control channel is not open")

// ErrorControlChannelInternalError is returned when an inconsistent state
// is encountered.
var ErrorControlChannelInternalError = errors.New("Internal Error")

type completion interface {
	encoding.TextUnmarshaler
}

type req interface {
	encoding.TextMarshaler
	Type() string
}

// A ControlChannelReader is a function that will fill its parameter with
// values after reading from some abstract pipe, returning nil or an error.
type ControlChannelReader func(interface{}) error

// A ControlChannelWriter is a function that will write its parameter to
// an abstract pipe, returning nil or an error.
type ControlChannelWriter func(interface{}) error

// ControlChannelState describes the current state of a ControlChannel.
type ControlChannelState uint

const (
	ControlChannelStateNew ControlChannelState = iota
	ControlChannelStateWaiting
	ControlChannelStateOpen
	ControlChannelStateClosing
	ControlChannelStateClosed
)

type issuedRequest uint64
type ControlChannel struct {
	m         sync.RWMutex
	last      issuedRequest
	pend      map[issuedRequest]chan completion
	state     ControlChannelState
	r         ControlChannelReader
	w         ControlChannelWriter
	lastError error
}

// RequestPayload encodes the packet format for a request.
type RequestPayload struct {
	T string
	P []byte
	I issuedRequest
}

// LastError returns the last error that occurred on this ControlChannel.
// Will return nil if everything's OK.
func (c *ControlChannel) LastError() error {
	return c.lastError
}

// State returns one of the ControlChannelState constants that describes
// this ControlChannel.
func (c *ControlChannel) State() ControlChannelState {
	return c.state
}

// IsOpen returns true if c.State() == ControlChannelStateOpen.
func (c *ControlChannel) IsOpen() bool {
	return c.state == ControlChannelStateOpen
}

func (c *ControlChannel) abortReq(r issuedRequest) {
	c.m.Lock()
	defer c.m.Unlock()
	ch, ok := c.pend[r]
	if !ok {
		return
	}
	delete(c.pend, r)
	if ch == nil {
		return
	}
	close(ch)
}

// Close will terminate the control channel and close any pending request
// chans.
func (c *ControlChannel) Close() {
	c.state = ControlChannelStateClosing
	defer func() { c.state = ControlChannelStateClosed }()

	entlen := len(c.pend)
	// If there are no pending requests, we're done
	if entlen == 0 {
		return
	}

	ents := make([]issuedRequest, 0, entlen)
	c.m.RLock()
	for k := range c.pend {
		ents = append(ents, k)
	}
	c.m.RUnlock()
	for _, r := range ents {
		c.abortReq(r)
	}
}

// fail sets c.lastError and calls c.Close().
func (c *ControlChannel) fail(e error) {
	c.lastError = e
	c.Close()
}

func (c *ControlChannel) issue(r *req) (compChan chan completion, err error) {
	if !c.IsOpen() {
		err = ErrorControlChannelNotOpen
		return
	}

	if c.w == nil || c.r == nil {
		err = ErrorControlChannelInternalError
		return
	}

	buf, err := (*r).MarshalText()
	if err != nil {
		return
	}

	compChan = make(chan completion)

	c.m.Lock()
	I := c.last + 1
	c.last = I
	c.pend[c.last] = compChan
	c.m.Unlock()

	p := RequestPayload{T: (*r).Type(), P: buf, I: I}
	err = c.w(p)
	if err != nil {
		c.fail(err)
		compChan = nil
	}
	return
}

func NewControlChannel(r ControlChannelReader, w ControlChannelWriter) (c *ControlChannel) {
	//	c = ControlChannel{r: r, w: w, pend: make()}
	return
}
