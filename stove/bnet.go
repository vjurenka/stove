package main

import (
	"encoding/binary"
	"fmt"
	"github.com/HearthSim/hs-proto/go"
	"github.com/golang/protobuf/proto"
	"hash/fnv"
	"net"
	"time"
)

const (
	CONN_HOST = "localhost"
	CONN_PORT = 1119
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Service struct {
	Name string
}

func (s Service) GetHashedName() uint32 {
	h := fnv.New32a()
	h.Write([]byte(s.Name))
	return h.Sum32()
}

type ConnectionService struct{ Service }
type AuthServerService struct{ Service }
type AuthClientService struct{ Service }

type session struct {
	conn net.Conn
}

func (session *session) handleRequest(packet []byte) int {
	headerSize := binary.BigEndian.Uint16(packet)
	headerData := packet[2 : 2+int(headerSize)]

	header := &hsproto.BnetProtocol_Header{}
	err := proto.Unmarshal(headerData, header)
	check(err)

	packetEnd := 2 + int(headerSize) + int(header.GetSize())
	bodyData := packet[2+headerSize : packetEnd]

	if header.GetServiceId() == 0 && header.GetMethodId() == 1 {
		body := &hsproto.BnetProtocolConnection_ConnectRequest{}
		err = proto.Unmarshal(bodyData, body)
		check(err)

		// register services
		connService := ConnectionService{Service{"bnet.protocol.connection.ConnectionService"}}
		authServerService := AuthServerService{Service{"bnet.protocol.authentication.AuthenticationServer"}}
		authClientService := AuthClientService{Service{"bnet.protocol.authentication.AuthenticationClient"}}
		fmt.Printf("connService=%d, authServerService=%d, authClientService=%d\n",
			connService.GetHashedName(),
			authServerService.GetHashedName(),
			authClientService.GetHashedName(),
		)

		bindRequest := body.GetBindRequest()
		// iterate
		for _, importedHash := range bindRequest.GetImportedServiceHash() {
			fmt.Printf("Client imports service %x\n", importedHash)
		}

		for _, export := range bindRequest.GetExportedService() {
			fmt.Printf("Client exports service id=%d, hash=%x\n", export.GetId(), export.GetHash())
		}

		timestamp := uint64(time.Now().UnixNano() / 1000)
		epoch := uint32(time.Now().Unix())

		resp := &hsproto.BnetProtocolConnection_ConnectResponse{
			ServerId: &hsproto.BnetProtocol_ProcessId{
				Label: proto.Uint32(3868510373),
				Epoch: proto.Uint32(epoch),
			},
			ClientId: &hsproto.BnetProtocol_ProcessId{
				Label: proto.Uint32(1255760),
				Epoch: proto.Uint32(epoch),
			},
			BindResponse: &hsproto.BnetProtocolConnection_BindResponse{
				ImportedServiceId: []uint32{1, 7, 8, 9, 6, 10, 11, 3, 2, 12, 5, 4}, //  5, 8, 1, 9, 3, 10, 11, 2, 12, 7, 13, 6, 14, 4, 15},
			},
			BindResult: proto.Uint32(0),
			ServerTime: proto.Uint64(timestamp),
		}

		data, err := proto.Marshal(resp)
		check(err)

		header := &hsproto.BnetProtocol_Header{
			ServiceId: proto.Uint32(254),
			MethodId:  proto.Uint32(1),
			Token:     proto.Uint32(0),
			Size:      proto.Uint32(uint32(len(data))),
		}

		session.writePacket(header, data)
	} else {
		fmt.Printf("unsupported: %d, %d\n", header.GetServiceId(), header.GetMethodId())
	}
	return packetEnd
}

func createSession(conn net.Conn) session {
	s := session{conn}

	return s
}

func (session *session) writePacket(head *hsproto.BnetProtocol_Header, body []byte) {
	headerData, err := proto.Marshal(head)
	check(err)
	outPacket := make([]byte, 2+len(headerData)+len(body))
	binary.BigEndian.PutUint16(outPacket, uint16(len(headerData)))
	copy(outPacket[2:], headerData)
	copy(outPacket[2+len(headerData):], body)
	written, err := session.conn.Write(outPacket)
	if written != len(outPacket) {
		fmt.Println("didn't write full packet, fixme")
	}
	check(err)
}

func (session *session) serve() {
	defer session.conn.Close()

	buf := make([]byte, 1024)
	idx := 0
	for {
		read, err := session.conn.Read(buf[idx:])
		check(err)
		totalProcessed := 0
		for read > totalProcessed {
			idx += read

			processed := session.handleRequest(buf[:idx])
			totalProcessed += processed
			idx -= processed
			if processed > 0 && totalProcessed < read {
				copy(buf[:len(buf)-processed], buf[processed:])
			} else {
				break
			}
		}
		if idx == len(buf) {
			buf = append(buf, make([]byte, len(buf)>>1)...)
		}
	}
}

func main() {
	// Listen for incoming connections
	hostname := fmt.Sprintf("%s:%d", CONN_HOST, CONN_PORT)
	tcpAddr, err := net.ResolveTCPAddr("tcp", hostname)
	check(err)
	sock, err := net.ListenTCP("tcp", tcpAddr)
	defer sock.Close()
	check(err)

	fmt.Printf("Listening on %s:%d ...\n", CONN_HOST, CONN_PORT)
	for {
		conn, err := sock.Accept()
		check(err)

		s := createSession(conn)
		go s.serve()
	}
}
