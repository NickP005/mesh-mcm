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
	// Read IP list from src_addr
	var ips []string
	for i := 0; i < int(m.recv_tx.Len[0]); i += 4 {
		ip := fmt.Sprintf("%d.%d.%d.%d", m.recv_tx.Src_addr[i], m.recv_tx.Src_addr[i+1], m.recv_tx.Src_addr[i+2], m.recv_tx.Src_addr[i+3])
		ips = append(ips, ip)
	}
	return ips, nil
}

// Resolve tag
func (m *SocketData) ResolveTag(tag []byte) (WotsAddress, error) {
	m.send_tx = NewTX(nil)
	m.send_tx.ID1 = m.recv_tx.ID1
	m.send_tx.ID2 = m.recv_tx.ID2

	// Create an empty WotsAddress and set the tag
	wots_addr := WotsAddressFromBytes([]byte{})
	wots_addr.SetTAG(tag)
	// Set the destination address
	m.send_tx.Dst_addr = wots_addr.Address
	// Send OP_RESOLVE
	err := m.SendOP(OP_RESOLVE)
	if err != nil {
		return WotsAddress{}, err
	}

	err = m.recvTX()
	if err != nil {
		return WotsAddress{}, err
	}

	// Check if opcode is OP_SEND_RESOLVE
	if m.recv_tx.Opcode[0] != byte(OP_RESOLVE) {
		return WotsAddress{}, (fmt.Errorf("opcode is not OP_RESOLVE"))
	}

	// Check if send total is one, else tag not found
	if m.recv_tx.Send_total[0] != 1 {
		return WotsAddress{}, (fmt.Errorf("tag not found"))
	}

	// Copy the address
	wots_addr = WotsAddressFromBytes(m.recv_tx.Dst_addr[:])

	// Set the amount
	wots_addr.SetAmountBytes(m.recv_tx.Change_total[:])

	return wots_addr, nil
}

// Get balance of a WotsAddress
func (m *SocketData) GetBalance(wots_addr WotsAddress) (uint64, error) {
	m.send_tx = NewTX(nil)
	m.send_tx.ID1 = m.recv_tx.ID1
	m.send_tx.ID2 = m.recv_tx.ID2

	// Set the destination address
	m.send_tx.Src_addr = wots_addr.Address

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

	// Change total should be 1
	if m.recv_tx.Change_total[0] != 1 {
		return 0, (fmt.Errorf("address not found"))
	}

	// Get the balance
	return binary.LittleEndian.Uint64(m.recv_tx.Send_total[:]), nil
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
	fmt.Println("File length:", len(file))
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
	copy(block_hash[:], m.recv_tx.Src_addr[:])
	return block_hash, nil
}
