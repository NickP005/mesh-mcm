package go_mcminterface

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

const (
	// Address and reference lengths
	ADDR_LEN      = 40 // Full Address length
	ADDR_REF_LEN  = 16 // Address Reference length
	ADDR_TAG_LEN  = 20 // Address Tag length
	ADDR_HASH_LEN = 20 // Address Hash length

	// WOTS+ related lengths
	WOTS_ADDR_LEN = 2208 // Full WOTS+ Address length
	WOTS_PK_LEN   = 2144 // WOTS+ Public Key length
	WOTS_SIG_LEN  = 2144 // WOTS+ Signature length
	WOTS_TAG_LEN  = 12   // WOTS+ Address Tag length

	// Transaction Data type codes
	TXDAT_MDST = 0x00 // Multi-Destination type transaction data

	// Transaction DSA type codes
	TXDSA_WOTS = 0x00 // WOTS+ DSA type transaction validation data
)

// MDST represents a Multi-Destination structure
type MDST struct {
	Tag    [ADDR_TAG_LEN]byte // Destination address tag
	Ref    [ADDR_REF_LEN]byte // Optional destination reference
	Amount [8]byte            // Destination send amount
}

func NewDSTFromString(tag string, ref string, amount uint64) MDST {
	dst := MDST{}
	// convert hex tag to bytes
	tag_bytes, _ := hex.DecodeString(tag)
	copy(dst.Tag[:], tag_bytes)

	copy(dst.Ref[:], ref) // TODO!!!

	binary.LittleEndian.PutUint64(dst.Amount[:], amount)
	return dst
}

// WOTSVAL represents a WOTS+ validation structure
type WOTSVAL struct {
	Signature [WOTS_SIG_LEN]byte    // WOTS+ Address signature
	PubSeed   [WOTS_PUBSEEDLEN]byte // WOTS+ Address Public Seed
	Adrs      [WOTS_ADDRLEN]byte    // WOTS+ Hash Function Address Scheme
}

// TXHDR represents a Transaction Header structure
type TXHDR struct {
	Options     [4]byte        // Transaction options
	SrcAddr     [ADDR_LEN]byte // Source address
	ChgAddr     [ADDR_LEN]byte // Change address
	SendTotal   [TXAMOUNT]byte // Total amount to send
	ChangeTotal [TXAMOUNT]byte // Total amount to change
	FeeTotal    [TXAMOUNT]byte // Total fee
	BlkToLive   [8]byte        // Block-to-live expiration
}

// Bytes returns the byte representation of the TXHDR
func (hdr *TXHDR) Bytes() []byte {
	var bytes []byte
	bytes = append(bytes, hdr.Options[:]...)
	bytes = append(bytes, hdr.SrcAddr[:]...)
	bytes = append(bytes, hdr.ChgAddr[:]...)
	bytes = append(bytes, hdr.SendTotal[:]...)
	bytes = append(bytes, hdr.ChangeTotal[:]...)
	bytes = append(bytes, hdr.FeeTotal[:]...)
	bytes = append(bytes, hdr.BlkToLive[:]...)
	return bytes
}

// TXDAT represents Transaction Data for various TX types
type TXDAT struct {
	Mdst []MDST // Multi-Destination array
}

// Bytes returns the byte representation of the TXDAT
func (dat *TXDAT) Bytes() []byte {
	var bytes []byte
	for _, dst := range dat.Mdst {
		bytes = append(bytes, dst.Tag[:]...)
		bytes = append(bytes, dst.Ref[:]...)
		bytes = append(bytes, dst.Amount[:]...)
	}
	return bytes
}

func TXDATFromBytes(bytes []byte, many uint8) TXDAT {
	dat := TXDAT{}
	for i := 0; i < int(many); i++ {
		var dst MDST
		shift := copy(dst.Tag[:], bytes[i*TXADDRLEN:i*TXADDRLEN+TXTAGLEN])
		shift += copy(dst.Ref[:], bytes[i*TXADDRLEN+shift:i*TXADDRLEN+shift+ADDR_REF_LEN])
		shift += copy(dst.Amount[:], bytes[i*TXADDRLEN+shift:i*TXADDRLEN+shift+TXAMOUNT])
		dat.Mdst = append(dat.Mdst, dst)
	}
	return dat
}

// TXDSA represents Transaction Validation data for various DSA types
type TXDSA struct {
	Wots WOTSVAL // WOTS+ DSA validation data
}

// Bytes returns the byte representation of the TXDSA
func (dsa *TXDSA) Bytes() []byte {
	var bytes []byte
	bytes = append(bytes, dsa.Wots.Signature[:]...)
	bytes = append(bytes, dsa.Wots.PubSeed[:]...)
	bytes = append(bytes, dsa.Wots.Adrs[:]...)
	return bytes
}

// TXTLR represents a Transaction Trailer structure
type TXTLR struct {
	Nonce [8]byte       // Transaction nonce
	ID    [HASHLEN]byte // Transaction ID
}

// Bytes returns the byte representation of the TXTLR
func (tlr *TXTLR) Bytes() []byte {
	var bytes []byte
	bytes = append(bytes, tlr.Nonce[:]...)
	bytes = append(bytes, tlr.ID[:]...)
	return bytes
}

// TXENTRY represents a complete transaction entry
type TXENTRY struct {
	Hdr TXHDR
	Dat TXDAT
	Dsa TXDSA
	Tlr TXTLR
}

// NewTXENTRY creates and initializes a new TXENTRY
func NewTXENTRY() TXENTRY {
	tx := TXENTRY{}
	tx.Hdr.Options[0] = TXDAT_MDST
	tx.Hdr.Options[1] = TXDSA_WOTS
	return tx
}

func TransactionFromHex(tx_hex string) TXENTRY {
	bytes, _ := hex.DecodeString(tx_hex)
	/* POSTPONE TO FUTURE ME LENGHT CHECK BY READING OPTIONS
	if len(bytes) == 0 {
		return TXENTRY{}
	}*/
	return TransactionFromBytes(bytes)
}

func TransactionFromBytes(bytes []byte) TXENTRY {
	tx := NewTXENTRY()

	shift := copy(tx.Hdr.Options[:], bytes[:4])
	shift += copy(tx.Hdr.SrcAddr[:], bytes[shift:shift+TXADDRLEN])
	shift += copy(tx.Hdr.ChgAddr[:], bytes[shift:shift+TXADDRLEN])
	shift += copy(tx.Hdr.SendTotal[:], bytes[shift:shift+TXAMOUNT])
	shift += copy(tx.Hdr.ChangeTotal[:], bytes[shift:shift+TXAMOUNT])
	shift += copy(tx.Hdr.FeeTotal[:], bytes[shift:shift+TXAMOUNT])
	shift += copy(tx.Hdr.BlkToLive[:], bytes[shift:shift+8])

	many_dst := tx.Hdr.Options[2] + 1
	tx.Dat = TXDATFromBytes(bytes[shift:], many_dst)
	shift += int(many_dst) * (TXADDRLEN + ADDR_REF_LEN + TXAMOUNT)

	shift += copy(tx.Dsa.Wots.Signature[:], bytes[shift:shift+WOTS_SIG_LEN])
	shift += copy(tx.Dsa.Wots.PubSeed[:], bytes[shift:shift+WOTS_PUBSEEDLEN])
	shift += copy(tx.Dsa.Wots.Adrs[:], bytes[shift:shift+WOTS_ADDR_LEN])

	shift += copy(tx.Tlr.Nonce[:], bytes[shift:shift+8])
	shift += copy(tx.Tlr.ID[:], bytes[shift:shift+HASHLEN])

	return tx
}

func (Transaction *TXENTRY) String() string {
	return hex.EncodeToString(Transaction.Bytes())
}

func (Transaction *TXENTRY) GetSignatureScheme() string {
	switch Transaction.Hdr.Options[1] {
	case TXDSA_WOTS:
		return "wotsp"
	default:
		return "unknown"
	}
}

func (Transaction *TXENTRY) SetSignatureScheme(scheme string) {
	switch scheme {
	case "wotsp":
		Transaction.Hdr.Options[1] = TXDSA_WOTS
	default:
		Transaction.Hdr.Options[1] = 0
	}
}

func (Transaction *TXENTRY) GetDestinationCount() uint8 {
	return Transaction.Hdr.Options[2] + 1
}

func (Transaction *TXENTRY) SetDestinationCount(count uint8) {
	Transaction.Hdr.Options[2] = count - 1
}

func (Transaction *TXENTRY) GetSourceAddress() WotsAddress {
	return WotsAddressFromBytes(Transaction.Hdr.SrcAddr[:])
}

func (Transaction *TXENTRY) SetSourceAddress(address WotsAddress) {
	copy(Transaction.Hdr.SrcAddr[:], address.Bytes())
}

func (Transaction *TXENTRY) GetChangeAddress() WotsAddress {
	return WotsAddressFromBytes(Transaction.Hdr.ChgAddr[:])
}

func (Transaction *TXENTRY) SetChangeAddress(address WotsAddress) {
	copy(Transaction.Hdr.ChgAddr[:], address.Bytes())
}

func (Transaction *TXENTRY) GetSendTotal() uint64 {
	return binary.LittleEndian.Uint64(Transaction.Hdr.SendTotal[:])
}

func (Transaction *TXENTRY) SetSendTotal(total uint64) {
	binary.LittleEndian.PutUint64(Transaction.Hdr.SendTotal[:], total)
}

func (Transaction *TXENTRY) GetChangeTotal() uint64 {
	return binary.LittleEndian.Uint64(Transaction.Hdr.ChangeTotal[:])
}

func (Transaction *TXENTRY) SetChangeTotal(total uint64) {
	binary.LittleEndian.PutUint64(Transaction.Hdr.ChangeTotal[:], total)
}

func (Transaction *TXENTRY) GetFee() uint64 {
	return binary.LittleEndian.Uint64(Transaction.Hdr.FeeTotal[:])
}

func (Transaction *TXENTRY) SetFee(fee uint64) {
	binary.LittleEndian.PutUint64(Transaction.Hdr.FeeTotal[:], fee)
}

func (Transaction *TXENTRY) GetBlockToLive() uint64 {
	return binary.LittleEndian.Uint64(Transaction.Hdr.BlkToLive[:])
}

func (Transaction *TXENTRY) SetBlockToLive(blk uint64) {
	binary.LittleEndian.PutUint64(Transaction.Hdr.BlkToLive[:], blk)
}

func (Transaction *TXENTRY) GetDestination(index uint8) MDST {
	return Transaction.Dat.Mdst[index]
}

func (Transaction *TXENTRY) SetDestination(index uint8, dst MDST) {
	Transaction.Dat.Mdst[index] = dst
}

func (Transaction *TXENTRY) GetDestinations() []MDST {
	return Transaction.Dat.Mdst
}

func (Transaction *TXENTRY) AddDestination(dst MDST) {
	Transaction.Dat.Mdst = append(Transaction.Dat.Mdst, dst)
}

func (Transaction *TXENTRY) GetWotsSignature() []byte {
	return Transaction.Dsa.Wots.Signature[:]
}

func (Transaction *TXENTRY) SetWotsSignature(signature []byte) {
	copy(Transaction.Dsa.Wots.Signature[:], signature)
}

func (Transaction *TXENTRY) GetWotsSigPubSeed() []byte {
	return Transaction.Dsa.Wots.PubSeed[:]
}

func (Transaction *TXENTRY) SetWotsSigPubSeed(seed [WOTS_PUBSEEDLEN]byte) {
	copy(Transaction.Dsa.Wots.PubSeed[:], seed[:])
}

func (Transaction *TXENTRY) GetWotsSigAddresses() []byte {
	return Transaction.Dsa.Wots.Adrs[:]
}

func (Transaction *TXENTRY) SetWotsSigAddresses(addresses []byte) {
	copy(Transaction.Dsa.Wots.Adrs[:], addresses)
}

func (Transaction *TXENTRY) GetNonce() uint64 {
	return binary.LittleEndian.Uint64(Transaction.Tlr.Nonce[:])
}

func (Transaction *TXENTRY) SetNonce(nonce uint64) {
	binary.LittleEndian.PutUint64(Transaction.Tlr.Nonce[:], nonce)
}

func (Transaction *TXENTRY) GetID() []byte {
	return Transaction.Tlr.ID[:]
}

func (Transaction *TXENTRY) SetID(id []byte) {
	copy(Transaction.Tlr.ID[:], id)
}

func (Transaction *TXENTRY) Bytes() []byte {
	var bytes []byte
	bytes = append(bytes, Transaction.Hdr.Bytes()...)
	bytes = append(bytes, Transaction.Dat.Bytes()...)
	bytes = append(bytes, Transaction.Dsa.Bytes()...)
	bytes = append(bytes, Transaction.Tlr.Bytes()...)
	return bytes
}

// return sha256 of transaction bytes
func (Transaction *TXENTRY) Hash() []byte {
	hash := sha256.New()
	hash.Write(Transaction.Bytes()[:HASHLEN])
	return hash.Sum(nil)
}
