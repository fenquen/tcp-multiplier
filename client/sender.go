package client

import (
	"encoding/hex"
	"go.uber.org/zap"
	"net"
	"strconv"
	"tcp-multiplier/config"
	"tcp-multiplier/utils"
	"tcp-multiplier/zaplog"
)

type Sender interface {
	Start()
	Run()

	Interrupt()
	Close()
	IsClosed() bool

	GetSrcDataChan() chan [] byte

	SetConn2DestSvr(conn2DestSvr net.Conn)
	SetSrcDataChan(srcDataChan chan []byte)
	SetSwitcher(switcher chan bool)
}

func NewSender(destTcpSvrAddrStr string, mode string) (Sender, error) {
	var conn2DestSvr net.Conn
	var result Sender

	switch mode {
	case "tcp":
		// localClientAddr
		localClientAddr, err := net.ResolveTCPAddr(mode,
			config.LocalClientHost+":"+strconv.Itoa(int(utils.GetLocalTcpClientPort())))
		if nil != err {
			zaplog.LOGGER.Error("ResolveTCPAddr localClientAddr", zap.Any("err", err), zap.Any("LocalClientHost", config.LocalClientHost))
			return nil, err
		}

		// destSvrAddr
		destSvrAddr, err := net.ResolveTCPAddr(mode, destTcpSvrAddrStr)
		if nil != err {
			zaplog.LOGGER.Error("ResolveTCPAddr destSvrAddr", zap.Any("err", err), zap.Any("destTcpSvrAddrStr", destTcpSvrAddrStr))
			return nil, err
		}

		// conn2DestSvr
		conn2DestSvr, err = net.DialTCP(mode, localClientAddr, destSvrAddr)
		if nil != err {
			zaplog.LOGGER.Error("DialTCP conn2DestSvr", zap.Any("err", err), zap.Any("destTcpSvrAddrStr", destTcpSvrAddrStr))
			return nil, err
		}

		result = &TcpSender{}
	case "udp":
		// localClientAddr
		localClientAddr, err := net.ResolveUDPAddr(mode,
			config.LocalClientHost+":"+strconv.Itoa(int(utils.GetLocalTcpClientPort())))
		if nil != err {
			zaplog.LOGGER.Error("ResolveUDPAddr localClientAddr", zap.Any("err", err), zap.Any("LocalClientHost", "LocalClientHost"))
			return nil, err
		}

		// destSvrAddr
		destSvrAddr, err := net.ResolveUDPAddr(mode, destTcpSvrAddrStr)
		if nil != err {
			zaplog.LOGGER.Error("ResolveUDPAddr destSvrAddr", zap.Any("err", err), zap.Any("destTcpSvrAddrStr", destTcpSvrAddrStr))
			return nil, err
		}

		// conn2DestSvr
		conn2DestSvr, err = net.DialUDP(mode, localClientAddr, destSvrAddr)
		if nil != err {
			zaplog.LOGGER.Error("DialUDP conn2DestSvr", zap.Any("err", err), zap.Any("destTcpSvrAddrStr", destTcpSvrAddrStr))
			return nil, err
		}

		result = &UdpSender{}
	}

	result.SetConn2DestSvr(conn2DestSvr)
	result.SetSrcDataChan(make(chan []byte, 100))
	result.SetSwitcher(make(chan bool, 1))

	return result, nil

}

type SenderBase struct {
	conn2DestSvr net.Conn
	srcDataChan  chan []byte
	switcher     chan bool
	closed       bool
}

func (senderBase *SenderBase) Start() {
	go senderBase.Run()
}

func (senderBase *SenderBase) Run() {
	defer func() {
		recover()

		senderBase.Close();
	}()

	for {
		// whether need to be interrupted
		select {
		case v := <-senderBase.switcher:
			if v {
				return
			}
		default:
		}

		select {
		case byteSlice := <-senderBase.srcDataChan:
			_, err := senderBase.conn2DestSvr.Write(byteSlice)
			if nil != err {
				zaplog.LOGGER.Info("senderBase.conn2DestSvr.Write", zap.Any("err", err))
				return
			}
			zaplog.LOGGER.Info("successfully write data to dest " + hex.EncodeToString(byteSlice))
		}
	}
}

func (senderBase *SenderBase) Interrupt() {
	senderBase.switcher <- true
	senderBase.Close()
}

func (senderBase *SenderBase) Close() {
	senderBase.closed = true
	_ = senderBase.conn2DestSvr.Close()
	close(senderBase.srcDataChan)
	close(senderBase.switcher)
}

func (senderBase *SenderBase) IsClosed() bool {
	return senderBase.closed
}

func (senderBase *SenderBase) GetSrcDataChan() chan [] byte {
	return senderBase.srcDataChan
}

func (senderBase *SenderBase) SetConn2DestSvr(conn2DestSvr net.Conn) {
	senderBase.conn2DestSvr = conn2DestSvr
}

func (senderBase *SenderBase) SetSrcDataChan(srcDataChan chan []byte) {
	senderBase.srcDataChan = srcDataChan
}
func (senderBase *SenderBase) SetSwitcher(switcher chan bool) {
	senderBase.switcher = switcher
}
