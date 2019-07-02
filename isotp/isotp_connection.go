package isotp

import "time"

//import "fmt"
import "github.com/andrewarrow/go-uds/util"

type AnyConn interface {
	Empty_rxqueue()
	Empty_txqueue()
	Send(payload []byte)
	Wait_frame() []byte
	Send_and_grant_flow_request(payload []byte) []byte
}

type IsotpConnection struct {
	name           string
	mtu            int
	fromIsoTPQueue *util.Queue
	toIsoTPQueue   *util.Queue
	Stack          *Transport
	rxfn           func() (Message, bool)
	txfn           func(msg Message)
}

func NewIsotpConnection(rx, tx int64, rxfn func() (Message, bool),
	txfn func(msg Message)) *IsotpConnection {
	ic := IsotpConnection{}
	a := NewAddress(rx, tx)
	ic.Stack = NewTransport(a, rxfn, txfn)
	ic.rxfn = rxfn
	ic.txfn = txfn
	ic.fromIsoTPQueue = util.NewQueue()
	ic.toIsoTPQueue = util.NewQueue()
	return &ic
}

func (ic *IsotpConnection) Send_and_grant_flow_request(payload []byte) []byte {
	msg := ic.Stack.make_tx_msg(ic.Stack.address.txid, payload)
	ic.Stack.txfn(msg)
	// wait for flow request
	time.Sleep(1 * time.Second)
	msg = ic.Stack.make_flow_control(CONTINUE)
	ic.Stack.txfn(msg)
	// read flow
	flow := []byte{}
	t1 := time.Now().Unix()
	for {
		msg, _ := ic.Stack.rxfn()
		if ic.Stack.address.is_for_me(msg) == false {
			continue
		}
		flow = append(flow, msg.Payload...)

		if time.Now().Unix()-t1 > 5 {
			break
		}
	}

	return flow
}

func (ic *IsotpConnection) Open() {
	go func() {
		for {
			//fmt.Println("  [ml] toIsoTP")
			for {
				if ic.toIsoTPQueue.Len() == 0 {
					break
				}
				payload := ic.toIsoTPQueue.Get()
				ic.Stack.Send(payload)
			}

			ic.Stack.Process()

			//fmt.Println("  [ml] fromIsoTP")
			for {
				if ic.Stack.available() == false {
					break
				}
				stuff := ic.Stack.Recv()
				ic.fromIsoTPQueue.Put(stuff)
			}
			//fmt.Println("  [ml] sleep")
			time.Sleep(time.Millisecond * 10)
		}
	}()
}

func (ic *IsotpConnection) Empty_rxqueue() {
	ic.fromIsoTPQueue.Clear()
}
func (ic *IsotpConnection) Empty_txqueue() {
	ic.toIsoTPQueue.Clear()
}
func (ic *IsotpConnection) Send(payload []byte) {
	//msg := NewMessage(ic.Stack.address.rxid, payload)
	ic.toIsoTPQueue.Put(payload)
}
func (ic *IsotpConnection) Wait_frame() []byte {

	count := 0
	for {
		if ic.fromIsoTPQueue.Len() > 0 {
			m := ic.fromIsoTPQueue.Get()
			return m
		}
		time.Sleep(500 * time.Millisecond)
		count++
		if count > 30 {
			break
		}
	}
	return []byte{}
}
