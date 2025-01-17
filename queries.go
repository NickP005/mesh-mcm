package go_mcminterface

import (
	"encoding/binary"
	"fmt"
)

// Get IP list
func (m *SocketData) GetIPList() ([]string, error) {
	// Send OP_GET_IPL
	err := m.SendOP(OP_GET_IPL)
	if err != nil {
		return nil, err
	}
	// Receive TX struct
	err = m.recvTX()
	if err != nil {
		return nil, err
	}
	// Check if opcode is OP_SEND_IPL
	if m.recv_tx.Opcode[0] != byte(OP_SEND_IPL) {
		fmt.Println("Opcode:", m.recv_tx.Opcode)
		return nil, (fmt.Errorf("opcode is not OP_SEND_IPL"))
	}

	// Read IP list from Buffer using buffer length
	bufferLen := binary.LittleEndian.Uint16(m.recv_tx.Len[:])
	var ips []string
	for i := 0; i < int(bufferLen); i += 4 {
		ip := fmt.Sprintf("%d.%d.%d.%d",
			m.recv_tx.Buffer[i],
			m.recv_tx.Buffer[i+1],
			m.recv_tx.Buffer[i+2],
			m.recv_tx.Buffer[i+3])
		ips = append(ips, ip)
	}
	return ips, nil
}

// Resolve tag
func (m *SocketData) ResolveTag(tag []byte) (WotsAddress, error) {
	m.send_tx = NewTX(nil)
	m.send_tx.ID1 = m.recv_tx.ID1
	m.send_tx.ID2 = m.recv_tx.ID2

	// Send an OP_BALANCE and retrieve the tag from the address
	m.send_tx.Buffer = tag
	binary.LittleEndian.PutUint16(m.send_tx.Len[:], uint16(len(tag)))

	// Send OP_BALANCE
	err := m.SendOP(OP_BALANCE)
	if err != nil {
		return WotsAddress{}, err
	}

	err = m.recvTX()
	if err != nil {
		return WotsAddress{}, err
	}

	// Check if opcode is OP_SEND_BALANCE
	if m.recv_tx.Opcode[0] != byte(OP_SEND_BAL) {
		return WotsAddress{}, (fmt.Errorf("opcode is not OP_SEND_BAL"))
	}

	// Check if the length is ADDR_LEN + TXAMOUNT
	var len uint16 = binary.LittleEndian.Uint16(m.recv_tx.Len[:])
	if len != ADDR_LEN+TXAMOUNT {
		return WotsAddress{}, (fmt.Errorf("length is not ADDR_LEN + AMOUNT_LEN"))
	}

	// Get the WotsAddress
	var wots_addr WotsAddress = WotsAddressFromBytes(m.recv_tx.Buffer[:])

	//fmt.Println("WotsAddress:", wots_addr)

	return wots_addr, nil
}

// Get balance of a WotsAddress
func (m *SocketData) GetBalance(wots_addr WotsAddress) (uint64, error) {
	//fmt.Println("GetBalance")
	m.send_tx = NewTX(nil)
	m.send_tx.ID1 = m.recv_tx.ID1
	m.send_tx.ID2 = m.recv_tx.ID2

	// Set the destination address
	m.send_tx.Buffer = wots_addr.Address[:]
	binary.LittleEndian.PutUint16(m.send_tx.Len[:], uint16(ADDR_LEN))

	//fmt.Println("Address:", wots_addr.Address)

	// Send OP_GET_BALANCE
	err := m.SendOP(OP_BALANCE)
	if err != nil {
		return 0, err
	}

	err = m.recvTX()
	if err != nil {
		return 0, err
	}

	// Check if opcode is OP_SEND_BALANCE
	if m.recv_tx.Opcode[0] != byte(OP_SEND_BAL) {
		return 0, (fmt.Errorf("opcode is not OP_SEND_BAL"))
	}

	var len uint16 = binary.LittleEndian.Uint16(m.recv_tx.Len[:])
	if len != ADDR_LEN+TXAMOUNT && len != ADDR_LEN {
		return 0, (fmt.Errorf("length is not ADDR_LEN + AMOUNT_LEN or ADDR_LEN"))
	}

	var wots_addr_recv WotsAddress = WotsAddressFromBytes(m.recv_tx.Buffer[:])

	//fmt.Println("Balance:", wots_addr_recv.Amount)

	return wots_addr_recv.Amount, nil
}

// Get block from block number
func (m *SocketData) GetBlockBytes(block_num uint64) ([]byte, error) {
	m.send_tx = NewTX(nil)
	m.send_tx.ID1 = m.recv_tx.ID1
	m.send_tx.ID2 = m.recv_tx.ID2

	// Set the block number
	binary.LittleEndian.PutUint64(m.send_tx.Blocknum[:], block_num)

	// Send OP_GET_BLOCK
	err := m.SendOP(OP_GET_BLOCK)
	if err != nil {
		return nil, err
	}

	file, err := m.recvFile()
	if err != nil {
		return nil, err
	}
	//print file length
	//fmt.Println("File length:", len(file))
	return file, nil
}

// Get block 32 bytes hash
func (m *SocketData) GetBlockHash(block_num uint64) ([HASHLEN]byte, error) {
	m.send_tx = NewTX(nil)
	m.send_tx.ID1 = m.recv_tx.ID1
	m.send_tx.ID2 = m.recv_tx.ID2

	if block_num == 0 {
		// set to the node's latest block number
		binary.LittleEndian.PutUint64(m.send_tx.Blocknum[:], m.block_num)
	} else {
		binary.LittleEndian.PutUint64(m.send_tx.Blocknum[:], block_num)
	}

	// Send OP_HASH
	err := m.SendOP(OP_HASH)
	if err != nil {
		return [HASHLEN]byte{}, err
	}

	err = m.recvTX()
	if err != nil {
		return [HASHLEN]byte{}, err
	}

	// Check if opcode is OP_HASH
	if m.recv_tx.Opcode[0] != byte(OP_HASH) {
		return [HASHLEN]byte{}, (fmt.Errorf("opcode is not OP_HASH"))
	}

	// Get the block hash
	var block_hash [HASHLEN]byte
	copy(block_hash[:], m.recv_tx.Buffer[:])
	return block_hash, nil
}

func (m *SocketData) GetTrailersBytes(block_num uint32, count uint32) ([]byte, error) {
	m.send_tx = NewTX(nil)
	m.send_tx.ID1 = m.recv_tx.ID1
	m.send_tx.ID2 = m.recv_tx.ID2

	// refer to specifications
	if count > 1000 {
		return nil, (fmt.Errorf("count is too high"))
	}

	// first 4 bytes of block  number are block_num, next 4 bytes are count
	binary.LittleEndian.PutUint32(m.send_tx.Blocknum[:4], block_num)
	binary.LittleEndian.PutUint32(m.send_tx.Blocknum[4:], count)

	// Send OP_TF
	err := m.SendOP(OP_TF)
	if err != nil {
		return nil, err
	}

	// receive file
	file, err := m.recvFile()
	if err != nil {
		return nil, err
	}

	// check if file length is correct
	if len(file)%160 != 0 {
		return nil, (fmt.Errorf("file length is not a multiple of 160"))
	}

	// iterate over the file and create trailers
	//for i := 0; i < len(file); i += 160 {
	//trailer := bTrailerFromBytes(file[i : i+160])
	//trailers = append(trailers, trailer)
	//}

	return file, nil
}

// Submit a transaction
func (m *SocketData) SubmitTransaction(tx TXENTRY) error {
	m.send_tx = NewTX(nil)
	m.send_tx.ID1 = m.recv_tx.ID1
	m.send_tx.ID2 = m.recv_tx.ID2

	// Set the transaction
	tx_entry := tx.Bytes()
	tx_entry_len := len(tx_entry)

	// Set the transaction length
	binary.LittleEndian.PutUint16(m.send_tx.Len[:], uint16(tx_entry_len))

	// Set the transaction
	m.send_tx.Buffer = tx_entry

	// Send OP_TX
	err := m.SendOP(OP_TX)
	if err != nil {
		return err
	}

	return nil
}
