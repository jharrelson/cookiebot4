package main

import (
	"net"
	"fmt"
)

type BncsPacket struct {
	Packet
}

type BnlsPacket struct {
	Packet
}

type Packet struct {
	position int
	buf []byte
}

func NewBnlsPacket(data []byte) (bnls *BnlsPacket) {
	bnls = new (BnlsPacket)
	if data == nil {
		bnls.buf = make([]byte, 3, 4096)
	} else {
		bnls.buf = append(bnls.buf, data...)
	}

	return bnls
}

func NewBncsPacket(data []byte) (bncs *BncsPacket) {
	bncs = new(BncsPacket)
	if data == nil {
		bncs.buf = make([]byte, 4, 4096)
	} else {
		bncs.buf = append(bncs.buf, data...)
	}

	return bncs
}

func (p *Packet) WriteByte(c byte) {
	p.buf = append(p.buf, c)
}

func (p *Packet) WriteByteArray(b []byte) {
	for i := 0; i < len(b); i++ {
		p.buf = append(p.buf, b[i])
	}
}

func (p *Packet) WriteWord(word int) {
	p.buf = append(p.buf, 
		byte(word),
		byte(word >> 8))
}

func (p *Packet) WriteDword(dword int) {
	p.buf = append(p.buf,
		byte(dword),
		byte(dword >> 8),
		byte(dword >> 16),
		byte(dword >> 24))
}

func (p *Packet) WriteString(s string) {
	p.buf = append(p.buf, s...)
	p.buf = append(p.buf, 0)
}

func (p *Packet) ReadByte() (c byte) {
	c = p.buf[p.position]
	p.position++
	return c
}

func (p *Packet) ReadWord() (word int) {
	word = int(p.buf[p.position]) | int(p.buf[p.position+1] << 8)
	p.position += 2
	return word
}

func (p *Packet) ReadDword() (dword int) {
	dword = int(p.buf[p.position]) | int(p.buf[p.position+1]) << 8 | 
		int(p.buf[p.position+2]) << 16 | int(p.buf[p.position+3]) << 24
	p.position += 4
	return dword
}

func (p *Packet) ReadString() (s string) {
	for ; p.position < len(p.buf); p.position++ {
		if p.buf[p.position] == 0x00 {
			p.position++
			break
		}
		s += string(p.buf[p.position])
	}
	return s
}

func (p *Packet) ReadByteArray() (b []byte) {
	b = make([]byte, 0, 256)
	for ; p.position < len(p.buf); p.position++ {
		b = append(b, p.buf[p.position])
		if p.buf[p.position] == 0x00 {
			break
		}
	}
	return b
}

func (p *Packet) Dump() {
	fmt.Printf("buffer dump @ %p [data @ %p] (packet length: %d)\n", p, p.buf, len(p.buf))
	
	for i := 0; i < len(p.buf); i += 16 {
		fmt.Printf("%04x: ", i)

		for j := 0; j < 16; j++ {
			if j == 8 {
				fmt.Printf(" ")
			}

			if (i + j) < len(p.buf) {
				fmt.Printf("%02x ", p.buf[i+j])
			} else {
				fmt.Printf("   ")
			}
		}
		fmt.Printf("| ")

		for j := 0; j < 16; j++ {
			if j == 8 {
				fmt.Printf(" ")
			}

			if (i + j) < len(p.buf) {
				if p.buf[i+j] <= 0x20 {
					fmt.Printf(".")
				} else if p.buf[i+j] >= 0x7f {
					fmt.Printf(".")
				} else {
					fmt.Printf("%c", p.buf[i+j])
				}
			}
		}
		fmt.Println("")
	}
}

func (bncs *BncsPacket) SendPacket(conn net.Conn, id byte) (err error) {
	bncs.buf[0] = 0xff
	bncs.buf[1] = id
	bncs.buf[2] = byte(len(bncs.buf))
	bncs.buf[3] = byte(len(bncs.buf) >> 8)

	_, err = conn.Write(bncs.buf)
//	bncs.Dump()
	return err
}

func (bnls *BnlsPacket) SendPacket(conn net.Conn, id byte) (err error) {
	bnls.buf[0] = byte(len(bnls.buf))
	bnls.buf[1] = byte(len(bnls.buf) >> 8)
	bnls.buf[2] = id

	_, err = conn.Write(bnls.buf)
//	bnls.Dump()
	return err
}