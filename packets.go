package libEmpaquette

import (
	"io"
	"log"
	"fmt"
	"bufio"
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

const (
	CONN_ACCEPTED = iota
	CONN_REFUSED_UN_PROT_VER
	CONN_REFUSED_ID_REJECTED
	CONN_REFUSED_SERV_UNAVAI
	CONN_REFUSED_BAD_USRPASS
	CONN_REFUSED_NOT_AUTHORI
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
	nameLen uint16
	name [4] byte
	lvl uint8
	connectFlags uint8
	keepAlive uint16
}

type connectPayload struct {
	len uint16
}

type connackVarHdr struct {
	SP uint8 `bitfield:"1"`
	ReturnCode uint8
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
	
	name := [...]byte{'M','Q','T','T'}
	varHdr.nameLen = uint16(len(name))
	varHdr.name = name
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

func ProcessPkt(b byte, r *bufio.Reader) {
	typ := extractType(b)
	switch typ {
	case CONNACK:
		processConnack(r)
	}
}

func extractType(b byte) byte {
	return (b >> 4) & 0xFF;
}

func processConnack(r *bufio.Reader) {
	fmt.Println("Proces CONNACK")
	len := decodeRemLen(r)
	fmt.Println("len = ", len)
	var hdr connackVarHdr
	err := structex.Decode(r, &hdr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("sp = ", hdr.SP)
	fmt.Println("return code = ", hdr.ReturnCode)
}

func decodeRemLen(r *bufio.Reader) int {
	multiplier := 1
	value := 0
	for {
		encodedByte, err := r.ReadByte()
		if err != nil {
			log.Fatal(err)
		}
		value += int(encodedByte & 127) * multiplier
		multiplier *= 128
		if multiplier > 128*128*128 {
			log.Fatal("Malformed Remaining Length")
		}
		if (encodedByte & 128) == 0 {
			break
		}
	}
	return value
}
