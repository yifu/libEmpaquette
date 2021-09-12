package libEmpaquette

import (
	"io"
	"log"
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
	ctrl_pkt_type uint8 `bitfield:"4"`
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
	name [10]byte
}

func CreateConnect(w WriterByteWriter) error {
	var hdr fixedHdr
	hdr.flags = 0x00
	hdr.ctrl_pkt_type = CONNECT
	
	err := structex.Encode(w, hdr)
	if err != nil {
		return err
	}

	var varHdr connectVarHdr
	var payload connectPayload

	varLen := binary.Size(varHdr) + binary.Size(payload)
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
	
	name := [...]byte{'C','l','i','e','n','t','t','e','s','t'}
	payload.len = uint16(len(name))
	payload.name = name
	err = binary.Write(w, binary.BigEndian, payload)
	if err != nil {
		log.Fatal(err)
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
