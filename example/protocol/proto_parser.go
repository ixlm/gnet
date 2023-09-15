package protocol

import (
	"encoding/binary"
	"errors"
)

type ProtoParser interface {
	Encode([]byte, interface{}) ([]byte, error)
	Decode([]byte) (interface{}, error)
	GetPackLen([]byte) (int, error)
}

/*
header description
1个字节对齐

| uint32   |uint16||uint16|uint8|uint64|
|packagelen|magicnumber|msgid|mark|transid|  body|
*/
const (
	NetHeaderLen uint32 = 17 // 包头的长度
)

type NetHeader struct {
	PackLen     uint32 //包的总长度
	MagicNumber uint16 //魔术字
	MsgId       uint16 //消息id
	Mark        uint8  //mark
	SessId      uint64 //会话id
}

// 从byte转换成NetHeader
func NewHeaderFromBytes(buf []byte) (*NetHeader, error) {
	if cap(buf) < int(NetHeaderLen) {
		return nil, errors.New("invalid package header")
	}
	packLen := binary.BigEndian.Uint32(buf[:4])
	magicNumber := binary.BigEndian.Uint16(buf[4:6])
	msgId := binary.BigEndian.Uint16(buf[6:8])
	mark := uint8(buf[8])
	sessId := binary.BigEndian.Uint64(buf[9:])
	return &NetHeader{PackLen: packLen, MagicNumber: magicNumber, MsgId: msgId, Mark: mark, SessId: sessId}, nil
}

//网络字节序是大端

// NetHeader转换成Bytes
func (h *NetHeader) Bytes() ([]byte, error) {
	buf := make([]byte, 17)
	binary.BigEndian.PutUint32(buf, h.PackLen)
	binary.BigEndian.PutUint16(buf[4:], h.MagicNumber) //+4
	binary.BigEndian.PutUint16(buf[6:], h.MsgId)       //+4+2
	buf[8] = h.Mark                                    //4+2+2
	binary.BigEndian.PutUint64(buf[9:], h.SessId)      //4+2+2+1
	return buf, nil
}

// 写入buffer
func (h *NetHeader) WriteTo(buf []byte) (int32, error) {
	if cap(buf) < int(NetHeaderLen) {
		return 0, errors.New("buf is not enough to hold the data")
	}
	binary.BigEndian.PutUint32(buf, h.PackLen)
	binary.BigEndian.PutUint16(buf[4:], h.MagicNumber) //+4
	binary.BigEndian.PutUint16(buf[6:], h.MsgId)       //+4+2
	buf[8] = h.Mark                                    //4+2+2
	binary.BigEndian.PutUint64(buf[9:], h.SessId)      //4+2+2+1
	return int32(NetHeaderLen), nil
}

// 判断2个包头是否相等
func (h *NetHeader) Equal(val *NetHeader) bool {
	return h.PackLen == val.PackLen && h.MagicNumber == val.MagicNumber && h.MsgId == val.MsgId && h.Mark == val.Mark && h.SessId == val.SessId
}

// 定义了一个自定义的解析器
type CustomParser struct{}

// -------------------- ProtoParser interface begin -------------------
func (parser *CustomParser) Encode(buf []byte, info interface{}) ([]byte, error) {
	bufLen := len(buf)
	msgLen := NetHeaderLen + uint32(bufLen)
	header := info.(*NetHeader)
	header.PackLen = msgLen
	data := make([]byte, msgLen)
	if ret, err := header.WriteTo(data); err != nil || ret != int32(NetHeaderLen) {
		return nil, errors.New("write to buffer failed!")
	}
	if ret := copy(data[msgLen:], buf); ret != bufLen {
		return nil, errors.New("copy data failed!")
	}
	return data, nil
}

func (parser *CustomParser) GetPackLen(buf []byte) (uint32, error) {
	bufLen := len(buf)
	if bufLen < int(NetHeaderLen) {
		return 0, errors.New("invalid package")
	}
	return binary.BigEndian.Uint32(buf), nil
}

func (parser *CustomParser) Decode(buf []byte) (interface{}, error) {
	// bufLen := len(buf)
	// if bufLen < int(NetHeaderLen) {
	// 	return nil, errors.New("invalid package")
	// }
	h, err := NewHeaderFromBytes(buf)
	if err != nil {
		return nil, errors.New("parse header failed")
	}
	// bodyLen := h.PackLen - NetHeaderLen

	return h, nil
}

//-------------------- ProtoParser interface end -------------------
