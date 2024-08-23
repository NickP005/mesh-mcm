package go_mcminterface

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/sigurn/crc16"
)

// Settings
const (
	SOCK_READ_TIMEOUT  = 5
	SOCK_WRITE_TIMEOUT = 5
	DEFAULT_PORT       = 2094
)

// Define constants
const (
	PVERSION  = 4      /* protocol version number (short) */
	CWALLET   = 2      /* indicate that we are a wallet */
	TXNETWORK = 0x3905 /* network number for transactions */
	TXTRAILER = 0xcdab /* trailer for transactions comms */

	OP_NULL       = 0  /* null operation code */
	OP_HELLO      = 1  /* hello first step in handshake */
	OP_HELLO_ACK  = 2  /* hello acknowledgement second step in handshake */
	FIRST_OP      = 3  /* first valid opcode after handshake */
	OP_TX         = 3  /* transaction opcode */
	OP_FOUND      = 4  /* block found opcode */
	OP_GET_BLOCK  = 5  /* get block opcode */
	OP_GET_IPL    = 6  /* get IP list opcode */
	OP_SEND_FILE  = 7  /* send file opcode */
	OP_SEND_IPL   = 8  /* send IP list opcode */
	OP_BUSY       = 9  /* busy opcode */
	OP_NACK       = 10 /* no acknowledge opcode */
	OP_GET_TFILE  = 11 /* get trailer file opcode */
	OP_BALANCE    = 12 /* get balance opcode */
	OP_SEND_BAL   = 13 /* send balance opcode */
	OP_RESOLVE    = 14 /* resolve tagged address opcode */
	OP_GET_CBLOCK = 15 /* get candidate block opcode */
	OP_MBLOCK     = 16 /* mined block opcode */
	OP_HASH       = 17 /* block hash opcode */
	OP_TF         = 18 /* get partial trailer file opcode */
	OP_IDENTIFY   = 19 /* identify opcode */
	LAST_OP       = 19 /* last valid opcode */

	TXADDRLEN    = 2208
	TXTAGLEN     = 12
	TXAMOUNT     = 8
	TXSIGLEN     = 2144
	HASHLEN      = 32
	BTRAILER_LEN = 160
)

type TX struct {
	Version      [2]byte
	Network      [2]byte
	ID1          [2]byte
	ID2          [2]byte
	Opcode       [2]byte
	Cblock       [8]byte
	Blocknum     [8]byte
	Cblockhash   [32]byte
	Pblockhash   [32]byte
	Weight       [32]byte
	Len          [2]byte
	Src_addr     [TXADDRLEN]byte
	Dst_addr     [TXADDRLEN]byte
	Chg_addr     [TXADDRLEN]byte
	Send_total   [TXAMOUNT]byte
	Change_total [TXAMOUNT]byte
	Tx_fee       [TXAMOUNT]byte
	Tx_sig       [TXSIGLEN]byte
	Crc16        [2]byte
	Trailer      [2]byte
}

// initialize the TX struct
func (m *TX) Init() {
	m.Version[0] = byte(PVERSION)
	m.Network[0] = byte(TXNETWORK >> 8)
	m.Network[1] = byte(TXNETWORK & 0xff)
	m.Trailer[0] = byte(TXTRAILER >> 8)
	m.Trailer[1] = byte(TXTRAILER & 0xff)

	// set ID1 to random value for
	rand.Read(m.ID1[:])
}

// Deserialize the TX struct
func (m *TX) Deserialize(bytes []byte) {
	// Deserialize the TX struct
	copy(m.Version[:], bytes[0:2])
	copy(m.Network[:], bytes[2:4])
	copy(m.ID1[:], bytes[4:6])
	copy(m.ID2[:], bytes[6:8])
	copy(m.Opcode[:], bytes[8:10])
	copy(m.Cblock[:], bytes[10:18])
	copy(m.Blocknum[:], bytes[18:26])
	copy(m.Cblockhash[:], bytes[26:58])
	copy(m.Pblockhash[:], bytes[58:90])
	copy(m.Weight[:], bytes[90:122])
	copy(m.Len[:], bytes[122:124])
	copy(m.Src_addr[:], bytes[124:2332])
	copy(m.Dst_addr[:], bytes[2332:4540])
	copy(m.Chg_addr[:], bytes[4540:6748])
	copy(m.Send_total[:], bytes[6748:6756])
	copy(m.Change_total[:], bytes[6756:6764])
	copy(m.Tx_fee[:], bytes[6764:6772])
	copy(m.Tx_sig[:], bytes[6772:8918])
	copy(m.Crc16[:], bytes[8916:8918])
	copy(m.Trailer[:], bytes[8918:8920])
}

// Serialize the TX struct
func (m *TX) serialize() []byte {
	var buf []byte
	buf = append(buf, m.Version[:]...)
	buf = append(buf, m.Network[:]...)
	buf = append(buf, m.ID1[:]...)
	buf = append(buf, m.ID2[:]...)
	buf = append(buf, m.Opcode[:]...)
	buf = append(buf, m.Cblock[:]...)
	buf = append(buf, m.Blocknum[:]...)
	buf = append(buf, m.Cblockhash[:]...)
	buf = append(buf, m.Pblockhash[:]...)
	buf = append(buf, m.Weight[:]...)
	buf = append(buf, m.Len[:]...)
	buf = append(buf, m.Src_addr[:]...)
	buf = append(buf, m.Dst_addr[:]...)
	buf = append(buf, m.Chg_addr[:]...)
	buf = append(buf, m.Send_total[:]...)
	buf = append(buf, m.Change_total[:]...)
	buf = append(buf, m.Tx_fee[:]...)
	buf = append(buf, m.Tx_sig[:]...)
	buf = append(buf, m.Crc16[:]...)
	buf = append(buf, m.Trailer[:]...)
	return buf
}

// GetBytes
// TO REMOVE IT'S REDUNDANT
func (m *TX) GetBytes() []byte {
	//m.computeCRC16()
	return m.serialize()
}

func NewTX(bytes []byte) TX {
	var tx TX
	if bytes != nil {
		// Deserialize the TX struct
		tx.Deserialize(bytes)
	} else {
		// Create a new TX struct
		tx.Init()
	}
	return tx
}

// Compute the CRC16 checksum up to signature
func (m *TX) computeCRC16() {
	buf := m.serialize()[:8916]
	//fmt.Printf("buf: %x\n", buf)
	table := crc16.MakeTable(crc16.CRC16_XMODEM)
	rcrc16 := crc16.Checksum(buf, table)
	m.Crc16[0] = byte(rcrc16 & 0xff)
	m.Crc16[1] = byte(rcrc16 >> 8)
}

// Get the version of the MCM interface
func (m *TX) GetVersion() int {
	return int(m.Version[0])
}

type SocketData struct {
	IP        string
	Conn      net.Conn
	send_tx   TX
	recv_tx   TX
	block_num uint64
}

// Send OP to IP
func (m *SocketData) SendOP(op uint16) error {
	// Set the opcode
	m.send_tx.Opcode = [2]byte{byte(op & 0xff), byte(op >> 8)}
	m.send_tx.computeCRC16()
	//fmt.Println("Sending OP:", op)
	// Send the TX struct
	return m.sendTX()
}

// Connect to IP : 2095
func (m *SocketData) Connect() {
	// Connect to the IP
	// print
	address := fmt.Sprintf("%s:%d", m.IP, DEFAULT_PORT)
	fmt.Println("Connecting to:", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	fmt.Println("Connected to:", address)
	m.Conn = conn
	m.Conn.SetWriteDeadline(time.Now().Add(SOCK_WRITE_TIMEOUT * time.Second))
	m.Conn.SetReadDeadline(time.Now().Add(SOCK_READ_TIMEOUT * time.Second))
}

// Send TX struct to IP
func (m *SocketData) sendTX() error {
	// Check if connection is active
	if m.Conn == nil {
		return fmt.Errorf("connection is nil")
	}
	bytes := m.send_tx.GetBytes()
	_, err := m.Conn.Write(bytes)
	if err != nil {
		fmt.Println("Error writing:", err)
		return err
	}
	return nil
}

// Receive TX struct from IP
func (m *SocketData) recvTX() error {
	buf := make([]byte, 8920)
	// read full
	n, err := io.ReadFull(m.Conn, buf)
	if err != nil {
		if err == io.EOF && n != 0 {
			fmt.Println("Connection closed before reading all bytes")
		} else if n != 0 {
			fmt.Println("Error reading:", err)
		}
		return err
	}
	// If received less than 8920 bytes, return
	if n < 8920 {
		return fmt.Errorf("received less than 8920 bytes")
	}
	// Deserialize the TX struct
	m.recv_tx = NewTX(buf)
	// print received OPCODE
	//fmt.Println("Received OPCODE:", m.recv_tx.Opcode[0])

	table := crc16.MakeTable(crc16.CRC16_XMODEM)
	rcrc16 := crc16.Checksum(m.recv_tx.GetBytes()[:8916], table)
	// Check if rcrc16 is equal to crc16
	if rcrc16 != binary.LittleEndian.Uint16(m.recv_tx.Crc16[:]) {
		fmt.Printf("crc16: %x\n", m.recv_tx.Crc16)
		fmt.Printf("rcrc16: %x\n", rcrc16)
		// print ID1
		fmt.Println("ID1:", m.recv_tx.ID1)
		// print ID2
		fmt.Println("ID2:", m.recv_tx.ID2)

		// print the received tx
		//fmt.Println("recv_tx:", m.recv_tx.GetBytes())
		return fmt.Errorf("crc16 checksum failed")
	}

	// Check the trailer
	if binary.BigEndian.Uint16(m.recv_tx.Trailer[:]) != TXTRAILER {
		return fmt.Errorf("trailer failed")
	}

	// Get the block number
	m.block_num = binary.LittleEndian.Uint64(m.recv_tx.Cblock[:])
	return nil
}

// Receive file from IP
func (m *SocketData) recvFile() ([]byte, error) {
	var file []byte

	// Until the connection is closed, keep receiving TX structs
	for {
		err := m.recvTX()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		// Check if opcode is OP_SEND_FILE
		if m.recv_tx.Opcode[0] != byte(OP_SEND_FILE) {
			return nil, fmt.Errorf("opcode is not OP_SEND_FILE")
		}

		// Bytes received in len
		len := binary.LittleEndian.Uint16(m.recv_tx.Len[:])

		// Get the bytes
		file = append(file, m.recv_tx.GetBytes()[124:124+len]...)
	}
	return file, nil
}

// Copy ID2 from recv_tx to send_tx
func (m *SocketData) copyID2() {
	//copy(m.recv_tx.ID2[:], m.send_tx.ID2[:])
	m.send_tx.ID2 = m.recv_tx.ID2
}

// Connnect and send OP_HELLO and wait for OP_HELLO_ACK
func (m *SocketData) Hello() error {
	// Connect to the IP
	m.Connect()
	// Send OP_HELLO
	err := m.SendOP(OP_HELLO)
	if err != nil {
		return err
	}
	//fmt.Println("Sent OP_HELLO")
	// Receive TX struct
	err = m.recvTX()
	//fmt.Println("Received TX struct")
	if err != nil {
		return err
	}
	// Check if opcode is OP_HELLO_ACK
	if m.recv_tx.Opcode[0] != byte(OP_HELLO_ACK) {
		return fmt.Errorf("opcode is not OP_HELLO_ACK")
	}
	// Copy ID2 from recv_tx to send_tx
	m.copyID2()
	return nil
}

func ConnectToNode(ip string) SocketData {
	var sd SocketData
	sd.IP = ip
	sd.send_tx = NewTX(nil)
	err := sd.Hello()
	if err != nil {
		fmt.Println("Error:", err)
	}
	return sd
}

// main function
func cmainn() {
	// Connect to node 35.212.41.137 195.181.241.89 192.168.1.70
	//sd := ConnectToNode("192.168.1.70")
	//test_dl_block()
	//test_resolve_balance()
	Settings = LoadSettings("settings.json")

	//BenchmarkNodes(1)
	fmt.Println("------------------")
	fmt.Println("Benchmark finished")
	fmt.Println("------------------")

	//test_query_balance()
	//test_hash_latest_block()
	//test_query_latest_block()
	//test_resolve_tag()
	test_latest_bnum()
	SaveSettings(Settings)
	/*
		Settings = LoadSettings()
		ExpandIPs()
		// print IPs
		fmt.Println("IPs:", Settings.IPs)
		BenchmarkNodes(10)

		nodes := PickNodes(5)
		// for each node print ip
		for _, node := range nodes {
			fmt.Print(node.IP, ":", node.Ping, " ")
		}
		fmt.Println("")
		nodes = PickNodes(5)
		// for each node print ip
		for _, node := range nodes {
			fmt.Print(node.IP, ":", node.Ping, " ")
		}

		SaveSettings(Settings)*/

}
