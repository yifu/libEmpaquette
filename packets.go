package libEmpaquette

import (
	"io"
	"log"
	"fmt"
	"bytes"
	"math"
	"bufio"
	"net"
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

type str string

func (s *str) Read(r io.Reader) error {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	len := binary.BigEndian.Uint16(buf)

	buf = make([]byte, len)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}

	*s = str(buf)
	return nil
} 

func (s str) Write(w io.Writer) error {
	dest := make([]byte, 2)
	if len(s) > math.MaxUint16 {
		return fmt.Errorf("String: %v > math.MaxUInt16.", len(s))
	}
	binary.BigEndian.PutUint16(dest, uint16(len(s)))
	w.Write(dest)
	w.Write([]byte(s))
	return nil
}

type fixedHdr struct {
	controlPacketType uint8
	dup uint8
	qos uint8
	retain uint8
	remainingLength int
}

func (fh *fixedHdr) Read(r io.Reader) error {
	buf := make([]byte, 1)
	_, err := r.Read(buf)
	if err != nil {
		return err
	}
	src := buf[0]
	flags := src & 0xF
	fh.retain = flags & 0x1
	flags >>= 1
	fh.qos = flags & 0x3
	flags >>= 2
	fh.dup = flags & 0x1
	fh.controlPacketType = (src >> 4) & 0xF
	fh.remainingLength, err = decodeRemLen(r)
	return err
}

func (fh fixedHdr) Write(w io.Writer) error {
	var dest uint8
	dest = (fh.controlPacketType & 0xF)
	
	dest <<= 1
	dest |= (fh.dup & 0x1)
	
	dest <<= 2
	dest |= (fh.qos & 0x3)
	
	dest <<= 1
	dest |= (fh.retain & 0x1)

	buf := make([]byte, 1)
	buf[0] = dest
	w.Write(buf)

	return encodeRemLen(w, fh.remainingLength)
}

type connectMsg struct {
	protocolName str
	protocolLevel uint8
	connectFlags uint8
	keepAlive uint16
	
	clientID str
}

func (m connectMsg) Write(w io.Writer) error {
	if err := m.protocolName.Write(w); err != nil {
		return err
	}
	
	protocolLevel := make([]byte, 1)
	protocolLevel[0] = m.protocolLevel
	if _, err := w.Write(protocolLevel); err != nil {
		return err
	}

	connectFlags := make([]byte, 1)
	connectFlags[0] = m.connectFlags
	if _, err := w.Write(connectFlags); err != nil {
		return err
	}

	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, m.keepAlive)
	if _, err := w.Write(buf); err != nil {
		return err
	}
	
	if err := m.clientID.Write(w); err != nil {
		return err
	}
	return nil
}

type connackMsg struct {
	SessionPresent uint8 
	ReturnCode uint8
}

func (m *connackMsg) Read(r io.Reader) error {
	buf := make([]byte, 2)
	total := 0
	for total != len(buf) {
		n, err := r.Read(buf)
		total += n
		if err != nil {
			return err
		}
	}
	m.SessionPresent = buf[0] & 0x1
	m.ReturnCode = buf[1]
	return nil
}

type publishMsg struct {
	topicName str
	pktID uint16
	payload string
}

func (pm *publishMsg) Read(r io.Reader, fh fixedHdr) error {
	if err := pm.topicName.Read(r); err != nil {
		return err
	}

	if fh.qos > 0 {
		buf := make([]byte, 2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return err
		}
		pm.pktID = binary.BigEndian.Uint16(buf)
	}

	payloadLen := fh.remainingLength
	payloadLen -= len(pm.topicName) - 2 /* sizeof hdr of pm.topicName */
	if fh.qos > 0 {
		payloadLen -= 2 /* sizeof pm.pktID */
	}

	buf := make([]byte, payloadLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	pm.payload = string(buf)

	return nil
}

func (pm publishMsg) Write(w io.Writer, fh fixedHdr) error {
	if err := pm.topicName.Write(w); err != nil {
		return err
	}

	if fh.qos > 0 {
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, pm.pktID)
		if _, err := w.Write(buf); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte(pm.payload)); err != nil {
		return err
	}

	return nil
}

type pubackMsg struct {
	pktID uint16
}

func (pm *pubackMsg) Read(r io.Reader) error {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	pm.pktID = binary.BigEndian.Uint16(buf)
	return nil
}

func (ctx *Context) SendConnect(clientID string) error {
	fmt.Println("Send CONNECT")
	var fh fixedHdr
	fh.dup = 0x00
	fh.qos = 0x00
	fh.retain = 0x00
	fh.controlPacketType = CONNECT
	fh.remainingLength = 0
	
	var msg connectMsg
	msg.protocolName = "MQTT"
	msg.protocolLevel = 4
	msg.connectFlags = 0x00
	msg.keepAlive = 3600
	msg.clientID = str(clientID)

	var body bytes.Buffer
	if err := msg.Write(&body); err != nil {
		log.Fatal(err)
	}
	fh.remainingLength = body.Len()
	if err := fh.Write(ctx.w); err != nil {
		log.Fatal(err)
	}
	if _, err := body.WriteTo(ctx.w); err != nil {
		log.Fatal(err)
	}
	ctx.w.Flush()
	return nil
}

func (ctx *Context) PublishMsg() error {
	fmt.Println("PUBLISH MSG")
	var fh fixedHdr
	fh.dup = 0x00
	fh.qos = 0x01
	fh.retain = 0x00
	fh.controlPacketType = PUBLISH
	fh.remainingLength = 0

	var msg publishMsg
	msg.pktID = 1
	msg.topicName = "topic/test"
	msg.payload = "payloadtest"

	var body bytes.Buffer
	if err := msg.Write(&body, fh); err != nil {
		log.Fatal(err)
	}
	fh.remainingLength = body.Len()
	if err := fh.Write(ctx.w); err != nil {
		log.Fatal(err)
	}
	if _, err := body.WriteTo(ctx.w); err != nil {
		log.Fatal(err)
	}
	ctx.w.Flush()
	return nil
}

func (ctx *Context) ProcessPkt() error {
	var fh fixedHdr
	if err := fh.Read(ctx.r); err != nil {
		return err
	}
	fmt.Println("fh.ctrlpkttype=", fh.controlPacketType)
	fmt.Println("fh.dup =", fh.dup)
	fmt.Println("fh.qos=", fh.qos)
	fmt.Println("fh.retain=", fh.retain)
	fmt.Println("rem len = ", fh.remainingLength)

	switch fh.controlPacketType {
	case CONNACK:
		if err := ctx.processConnack(); err != nil {
			return err
		}
	case PUBACK:
		if err := ctx.processPuback(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unprocessed type of message: %v.", fh.controlPacketType)
	}
	return nil
}

func extractType(b byte) byte {
	return (b >> 4) & 0xFF;
}

func (ctx *Context)processConnack() error {
	fmt.Println("Process CONNACK")

	var msg connackMsg
	if err := msg.Read(ctx.r); err != nil {
		return err
	}
	fmt.Println("session present =", msg.SessionPresent)
	fmt.Println("return code =", msg.ReturnCode)
	return nil
}

func (ctx *Context)processPuback() error {
	fmt.Println("Process PUBACK")

	var msg pubackMsg
	if err := msg.Read(ctx.r); err != nil {
		return err
	}
	fmt.Println("pktID =", msg.pktID)
	return nil
}

func encodeRemLen(w io.Writer, varLen int) error {
	if varLen == 0 {
		dest := make([]byte, 1)
		dest[0] = 0
		_, err := w.Write(dest)
		return err
	}

	for varLen > 0 {
		var encodedByte byte = byte(varLen % 128)
		varLen = varLen / 128
		if varLen > 0 {
			encodedByte = encodedByte | 128
		}
		dest := make([]byte, 1)
		dest[0] = encodedByte
		if _, err := w.Write(dest); err != nil {
			return err
		}
	}
	return nil
}

func decodeRemLen(r io.Reader) (int, error) {
	multiplier := 1
	value := 0
	for {
		buf := make([]byte, 1)
		_, err := r.Read(buf)
		if err != nil {
			return 0, err
		}
		encodedByte := buf[0]
		value += int(encodedByte & 127) * multiplier
		multiplier *= 128
		if multiplier > 128*128*128 {
			return 0, fmt.Errorf("Malformed Remaining Length")
		}
		if (encodedByte & 128) == 0 {
			break
		}
	}
	return value, nil
}

type Context struct {
	r *bufio.Reader
	w *bufio.Writer
}

func Connect(ipport string) (Context, error) {
	conn, err := net.Dial("tcp", ipport)
	if err != nil {
		return Context{}, err
	}

	return Context{r: bufio.NewReader(conn), w: bufio.NewWriter(conn)}, nil
}