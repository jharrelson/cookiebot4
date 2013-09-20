package main

import (
	"net"
	"fmt"
)

type BncsPacket struct {
	Packet
}

type Packet struct {
	id byte
	position int
	buf []byte
}

type PacketWriter interface {
	WriteByte(c byte)
	WriteWord(word int16)
	WriteDword(dword int32)
	WriteString(s string)
	SendPacket(conn net.Conn) (err error)
}

func NewBncsPacket(id byte) (bncs *BncsPacket) {
	bncs = new(BncsPacket)
	bncs.id = id
	bncs.buf = make([]byte, 4, 4096)

	return bncs
}

func (p *Packet) WriteByte(c byte) {
	p.buf = append(p.buf, c)
}

func (p *Packet) WriteWord(word int16) {
	p.buf = append(p.buf, 
		byte(word),
		byte(word >> 8))
}

func (p *Packet) WriteDword(dword uint32) {
	p.buf = append(p.buf,
		byte(dword),
		byte(dword >> 8),
		byte(dword >> 16),
		byte(dword >> 24),)
}

func (p *Packet) WriteString(s string) {
	p.buf = append(p.buf, s...)
	p.buf = append(p.buf, 0)
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

func (bncs *BncsPacket) SendPacket(conn net.Conn) (err error) {
	bncs.buf[0] = 0xff
	bncs.buf[1] = bncs.id
	bncs.buf[2] = byte(len(bncs.buf))
	bncs.buf[3] = byte(len(bncs.buf) >> 8)

	_, err = conn.Write(bncs.buf)
	return err
}