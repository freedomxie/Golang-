package byteio

import (
	"bytes"
	"encoding/binary"
	"log"
)

//二进制解包与组包工具，用于网络通信中的字节的处理

type ByteIo struct {
	ByteOrder binary.ByteOrder
	Buf       []byte
	Index     int
}

func (this *ByteIo) Set(buf []byte) {
	this.Buf = this.Buf[0:0]
	this.Index = 0
	this.copyBuffer(buf)
}

func (this *ByteIo) Get() []byte {
	return this.Buf
}

func (this *ByteIo) Clear() {
	this.Buf = this.Buf[0:0]
	this.Index = 0
}

func (this *ByteIo) Remaind() []byte {
	return this.Buf[this.Index:]
}

func (this *ByteIo) Wbytes(bytes []byte) {
	this.Buf = append(this.Buf, bytes...)
}

func (this *ByteIo) copyBuffer(buf []byte) {
	size := len(this.Buf)
	var length int = size + len(buf)
	newBuf := make([]byte, length)
	copy(newBuf, this.Buf)
	copy(newBuf[size:], buf)
	this.Buf = newBuf

}

func (this *ByteIo) Head2(buf []byte) []byte {
	length := len(buf)
	rBuf := make([]byte, length+2)

	newBuf := bytes.NewBuffer([]byte{})
	binary.Write(newBuf, this.ByteOrder, int16(length))
	copy(rBuf, newBuf.Bytes())
	copy(rBuf[2:], buf)
	return rBuf
}

func (this *ByteIo) Head4(buf []byte) []byte {
	length := len(buf)
	rBuf := make([]byte, length+4)

	newBuf := bytes.NewBuffer([]byte{})
	binary.Write(newBuf, this.ByteOrder, int32(length))
	copy(rBuf, newBuf.Bytes())
	copy(rBuf[4:], buf)
	return rBuf
}

func (this *ByteIo) Wbyte(v byte) {
	newBuf := bytes.NewBuffer([]byte{})
	binary.Write(newBuf, this.ByteOrder, v)
	this.copyBuffer(newBuf.Bytes())
}

func (this *ByteIo) Wint16(v int16) {
	newBuf := bytes.NewBuffer([]byte{})
	binary.Write(newBuf, this.ByteOrder, v)
	this.copyBuffer(newBuf.Bytes())
}

func (this *ByteIo) Wuint16(v uint16) {
	newBuf := bytes.NewBuffer([]byte{})
	binary.Write(newBuf, this.ByteOrder, v)
	this.copyBuffer(newBuf.Bytes())
}

func (this *ByteIo) Wint32(v int32) {
	newBuf := bytes.NewBuffer([]byte{})
	binary.Write(newBuf, this.ByteOrder, v)
	this.copyBuffer(newBuf.Bytes())
}

func (this *ByteIo) Wint64(v int64) {
	newBuf := bytes.NewBuffer([]byte{})
	binary.Write(newBuf, this.ByteOrder, v)
	this.copyBuffer(newBuf.Bytes())
}

func (this *ByteIo) Wstr(v string) {
	buf := []byte(v)
	buf = append(buf, '\000')

	newBuf := bytes.NewBuffer([]byte{})
	binary.Write(newBuf, this.ByteOrder, buf)
	this.copyBuffer(newBuf.Bytes())
}

func (this *ByteIo) WstrCtrlA(v string) {
	buf := []byte(v)
	buf = append(buf, '\001')

	newBuf := bytes.NewBuffer([]byte{})
	binary.Write(newBuf, this.ByteOrder, buf)
	this.copyBuffer(newBuf.Bytes())
}

func (this *ByteIo) Rbyte() byte {
	var x byte
	end := this.Index + 1
	bytesBuffer := bytes.NewBuffer(this.Buf[this.Index:end])
	binary.Read(bytesBuffer, this.ByteOrder, &x)
	this.Index += 1
	return x
}

func (this *ByteIo) Rint16() int16 {

	var x int16
	end := this.Index + 2
	bytesBuffer := bytes.NewBuffer(this.Buf[this.Index:end])
	binary.Read(bytesBuffer, this.ByteOrder, &x)
	this.Index += 2

	return int16(x)
}

func (this *ByteIo) Ruint16() uint16 {

	var x uint16
	end := this.Index + 2
	bytesBuffer := bytes.NewBuffer(this.Buf[this.Index:end])
	binary.Read(bytesBuffer, this.ByteOrder, &x)
	this.Index += 2

	return uint16(x)
}

func (this *ByteIo) Rint32() int32 {
	var x int32
	end := this.Index + 4
	bytesBuffer := bytes.NewBuffer(this.Buf[this.Index:end])
	binary.Read(bytesBuffer, this.ByteOrder, &x)
	this.Index += 4
	return int32(x)
}

func (this *ByteIo) Rint64() int64 {
	var x int64
	end := this.Index + 8
	bytesBuffer := bytes.NewBuffer(this.Buf[this.Index:end])
	binary.Read(bytesBuffer, this.ByteOrder, &x)
	this.Index += 8
	return int64(x)
}

func (this *ByteIo) RstrCtrlA() string {
	index := this.Index
	var end int
	for i := this.Index; i < len(this.Buf); i++ {
		if this.Buf[i] == '\001' {
			end = i
			break
		}
	}
	x := this.Buf[index:end]
	this.Index = end
	return string(x)
}

func (this *ByteIo) Rstr() string {
	index := this.Index
	var end int
	for i := this.Index; i < len(this.Buf); i++ {
		if this.Buf[i] == '\000' {
			end = i
			break
		}
	}

	x := this.Buf[index:end]
	this.Index = end
	return string(x)
}

func (this *ByteIo) RfixStr(length int) string {
	index := this.Index
	end := index + length
	log.Println("RfixStr:", "index:", index, "end:", end)
	x := this.Buf[index:end]
	this.Index = end
	return string(x)
}
