package vsocket_golang

import (
	"fmt"
	"testing"
	"time"
)

func TestCase1(t *testing.T) {
	s := SocketInfo{0, isAllowSend, send}
	Create(&s)
	RegisterObserver(rxCallback)

	go func() {
		rxParam := RxParam{}
		rxParam.Pipe = 0
		rxParam.Data = []uint8{1, 2, 3}
		rxParam.IP = 1
		rxParam.Port = 2
		Receive(&rxParam)
	}()

	time.After(time.Second * 2)
	txParam := TxParam{}
	txParam.Pipe = 0
	txParam.IP = 3
	txParam.Port = 4
	txParam.Data = []uint8{5, 6, 7}
	Send(&txParam)
	time.After(time.Second * 2)
}

func isAllowSend() bool {
	return true
}

func send(data []uint8, ip uint32, port uint16) {
	fmt.Println("send", ip, port, data)
}

func rxCallback(rxParam *RxParam) {
	fmt.Println("receive", rxParam)
}
