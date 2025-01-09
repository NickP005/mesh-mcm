package main

import (
	"encoding/binary"
	"encoding/hex"

	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
)

const (
	SHA3LEN512 = 64
)

type WotsAddress struct {
	Address [TXADDRLEN]byte
	Amount  uint64
}

func (m *WotsAddress) Bytes() []byte {
	var buf []byte
	buf = append(buf, m.Address[:]...)
	buf = append(buf, m.GetAmountBytes()...)
	return buf
}

func (m *WotsAddress) GetTAG() []byte {
	return m.Address[:ADDR_TAG_LEN]
}

func (m *WotsAddress) SetTAG(tag []byte) {
	copy(m.Address[:ADDR_TAG_LEN], tag)
}

func (m *WotsAddress) GetAddress() []byte {
	return m.Address[ADDR_TAG_LEN:]
}

func (m *WotsAddress) SetAddress(address []byte) {
	copy(m.Address[ADDR_TAG_LEN:], address)
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
	if len(bytes) == WOTS_PK_LEN {
		copy(wots.Address[:], AddrFromWots(bytes))
	} else if len(bytes) == TXADDRLEN {
		copy(wots.Address[:], bytes)
	} else if len(bytes) == TXADDRLEN+TXAMOUNT {
		copy(wots.Address[:], bytes[:TXADDRLEN])
		wots.SetAmountBytes(bytes[TXADDRLEN:])
	}

	return wots
}

func WotsAddressFromHex(wots_hex string) WotsAddress {
	bytes, _ := hex.DecodeString(wots_hex)
	if len(bytes) != TXADDRLEN {
		return WotsAddress{}
	}
	return WotsAddressFromBytes(bytes)
}

// AddrFromImplicit converts a tag to a full hash-based address
func AddrFromImplicit(tag []byte) []byte {
	addr := make([]byte, TXADDRLEN)
	// Copy tag to both tag and hash portions
	copy(addr[:ADDR_TAG_LEN], tag)
	copy(addr[ADDR_TAG_LEN:], tag)
	return addr
}

// AddrHashGenerate generates a Mochimo Address hash using SHA3-512 and RIPEMD160
func AddrHashGenerate(in []byte) []byte {
	// First pass: SHA3-512
	hash := make([]byte, SHA3LEN512)
	sha3Hash := sha3.New512()
	sha3Hash.Write(in)
	sha3Hash.Sum(hash[:0])

	// Second pass: RIPEMD160
	ripemd := ripemd160.New()
	ripemd.Write(hash)
	return ripemd.Sum(nil)
}

// AddrFromWots converts Legacy WOTS+ address to hash-based Mochimo Address
func AddrFromWots(wots []byte) []byte {
	if len(wots) != WOTS_PK_LEN {
		return nil
	}
	// Generate hash of WOTS public key
	hash := AddrHashGenerate(wots[:WOTS_PK_LEN])
	// Convert to implicit address
	return AddrFromImplicit(hash)
}
