package client

import (
	"net"
)

func Dial(addr string) (*net.UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func Write(conn *net.UDPConn, buf []byte) (int, error) {
	return conn.Write(buf)
}

func Read(conn *net.UDPConn, buf []byte) (int, error) {
	return conn.Read(buf)
}
