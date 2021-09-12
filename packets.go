package libEmpaquette

import (
	"io"
	"log"
	"fmt"
	"github.com/HewlettPackard/structex"
	"encoding/binary"
)
const (
	CONNECT = iota + 1
	CONNACK
	PUBLISH
	PUBACK
	PUBREC
	PUBREL
	PUBCOMP
	SUBSCRIBE
	SUBACK
	UNSUBSCRIBE
	UNSUBACK
	PINGREQ
	PINGRESP
	DISCONNECT
)

type WriterByteWriter interface {
	io.ByteWriter
	io.Writer
}

type fixedHdr struct {
	flags uint8 `bitfield:"4"`
	ctrlPktType uint8 `bitfield:"4"`
}

type connectVarHdr struct {
	len uint16
	name [4] byte
	lvl uint8
	connectFlags uint8
	keepAlive uint16
}

type connectPayload struct {
	len uint16
}

func CreateConnect(w WriterByteWriter, clientid string) error {
	var hdr fixedHdr
	hdr.flags = 0x00
	hdr.ctrlPktType = CONNECT
	
	err := structex.Encode(w, hdr)
	if err != nil {
		return err
	}

	var varHdr connectVarHdr
	var payload connectPayload

	varLen := binary.Size(varHdr) + binary.Size(payload) + len(clientid)
	fmt.Println("len(", clientid, ") = ", len(clientid))
	encodeLen(w, varLen)
	
	varHdr.len = 4
	varHdr.name = [...]byte{'M','Q','T','T'}
	varHdr.lvl = 4
	varHdr.connectFlags = 0x00
	varHdr.keepAlive = 3600
	err = binary.Write(w, binary.BigEndian, varHdr)
	if err != nil {
		log.Fatal(err)
	}
	
	payload.len = uint16(len(clientid))
	err = binary.Write(w, binary.BigEndian, payload)
	if err != nil {
		log.Fatal(err)
	}

	for _, b := range []byte(clientid) {
		err := w.WriteByte(b)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func encodeLen(w io.ByteWriter, varLen int) {
	if varLen == 0 {
		w.WriteByte(0)
		return
	}

	for varLen > 0 {
		var encodedByte byte = byte(varLen % 128)
		varLen = varLen / 128
		if varLen > 0 {
			encodedByte = encodedByte | 128
		}
		w.WriteByte(encodedByte)
	}
}
