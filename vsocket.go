// Copyright 2022-2022 The jdh99 Authors. All rights reserved.
// 虚拟端口
// Authors: jdh99 <jdh821@163.com>

package vsocket_golang

import (
	"github.com/jdhxyy/lagan"
	"time"
)

const (
	tag = "vsocket"

	fifoLen = 1000

	// 发送等待延时.单位:ms
	txWaitTime = 1
)

// IsAllowSendFunc 是否允许发送
type IsAllowSendFunc func() bool

// SendFunc 发送函数
type SendFunc func(data []uint8, ip uint32, port uint16)

// SocketInfo 端口信息
type SocketInfo struct {
	Pipe int

	// API
	IsAllowSend IsAllowSendFunc
	Send SendFunc
}

type tSocket struct {
	fifo chan *TxParam
	SocketInfo
}

// RxParam 端口接收数据参数
type RxParam struct {
	Pipe int
	Data []uint8
	Metric int
	// 接收时间.单位:us
	RxTime int64

	// 源节点网络信息.不是网络端口可不管这两个字段
	IP uint32
	Port uint16
}

// TxParam 端口发送数据参数
type TxParam struct {
	Pipe int
	Data []uint8

	// 源节点网络信息.不是网络端口可不管这两个字段
	IP uint32
	Port uint16
}

// RxCallback 接收回调函数
type RxCallback func(rxParam *RxParam)

var observers []RxCallback
var sockets map[int]*tSocket

func init() {
	lagan.Info(tag, "init")
	sockets = make(map[int]*tSocket)
}

// Create 创建socket
func Create(socketInfo *SocketInfo) {
	if socketInfo == nil {
		lagan.Error(tag, "socket create failed!param is wrong")
		return
	}

	lagan.Info(tag, "socket:%d create success", socketInfo.Pipe)
	s, ok := sockets[socketInfo.Pipe]
	if ok == false {
		sockets[socketInfo.Pipe] = &tSocket{make(chan *TxParam, fifoLen), *socketInfo}
		go checkTxFifo(sockets[socketInfo.Pipe])
	} else {
		s.SocketInfo = *socketInfo
	}
}

func checkTxFifo(s *tSocket) {
	for {
		t := <-s.fifo
		for {
			if s.IsAllowSend() == true {
				break
			}
			time.Sleep(time.Millisecond * txWaitTime)
		}
		s.Send(t.Data, t.IP, t.Port)
	}
}

// Send 发送数据.非网络端口不需要管ip和port两个字段
func Send(txParam *TxParam) {
	if txParam == nil {
		lagan.Error(tag, "send failed!param is nil")
		return
	}
	s, ok := sockets[txParam.Pipe]
	if ok == false {
		lagan.Error(tag, "send failed!can not find pipe:%d", txParam.Pipe)
		return
	}

	select {
	case s.fifo<-txParam:
	default:lagan.Error(tag, "%d send failed!fifo is full", s.Pipe)
	}
}

// Receive 接收数据.应用模块接收到数据后需调用本函数
// RxTime字段不需要管,填0即可
// 非网络端口不需要管ip和port两个字段,填0即可
func Receive(rxParam *RxParam) {
	if rxParam == nil {
		lagan.Error(tag, "receive failed!param is nil")
		return
	}
	_, ok := sockets[rxParam.Pipe]
	if ok == false {
		lagan.Error(tag, "receive failed!can not find pipe:%d", rxParam.Pipe)
		return
	}

	rxParam.RxTime = time.Now().UnixNano() / 1000
	notifyObservers(rxParam)
}

func notifyObservers(rxParam *RxParam) {
	n := len(observers)
	for i := 0; i < n; i++ {
		observers[i](rxParam)
	}
}

// RegisterObserver 注册观察者
func RegisterObserver(callback RxCallback) {
	observers = append(observers, callback)
}

// IsAllowSend 是否允许发送
func IsAllowSend(pipe int) bool {
	s, ok := sockets[pipe]
	if ok == false {
		return false
	}
	return s.IsAllowSend()
}
