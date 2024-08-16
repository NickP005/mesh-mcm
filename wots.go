package go_mcminterface

import (
	"encoding/binary"
	"encoding/hex"
)

type WotsAddress struct {
	Address [TXADDRLEN]byte
	Amount  uint64
}

func (m *WotsAddress) GetTAG() []byte {
	// return last 12 bytes of address
	return m.Address[TXADDRLEN-12:]
}

func (m *WotsAddress) SetTAG(tag []byte) {
	// set last 12 bytes of address
	copy(m.Address[TXADDRLEN-12:], tag)
}

func (m *WotsAddress) IsDefaultTag() bool {
	// check if tag is [66,0,0,0,14,0,0,0,1,0,0,0]
	tag := m.GetTAG()
	if tag[0] == 66 && tag[1] == 0 && tag[2] == 0 && tag[3] == 0 && tag[4] == 14 && tag[5] == 0 && tag[6] == 0 && tag[7] == 0 && tag[8] == 1 && tag[9] == 0 && tag[10] == 0 && tag[11] == 0 {
		return true
	}
	return false
}

func (m *WotsAddress) GetPublKey() []byte {
	// return first 2208 bytes of address
	return m.Address[:TXADDRLEN-12]
}

func (m *WotsAddress) SetPublKey(publKey []byte) {
	// set first 2208 bytes of address
	copy(m.Address[:TXADDRLEN-12], publKey)
}

func (m *WotsAddress) SetAmountBytes(amount []byte) {
	m.Amount = binary.LittleEndian.Uint64(amount)
}

func (m *WotsAddress) GetAmount() uint64 {
	return m.Amount
}

func (m *WotsAddress) GetAmountBytes() []byte {
	var amount [8]byte
	binary.LittleEndian.PutUint64(amount[:], m.Amount)
	return amount[:]
}

func WotsAddressFromBytes(bytes []byte) WotsAddress {
	var wots WotsAddress
	copy(wots.Address[:], bytes)
	return wots
}

func WotsAddressFromHex(wots_hex string) WotsAddress {
	bytes, _ := hex.DecodeString(wots_hex)
	if len(bytes) != TXADDRLEN {
		return WotsAddress{}
	}
	return WotsAddressFromBytes(bytes)
}
