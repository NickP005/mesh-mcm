package go_mcminterface

import "unsafe"

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

// WOTSVAL represents a WOTS+ validation structure
type WOTSVAL struct {
	Signature [WOTS_SIG_LEN]byte // WOTS+ Address signature
	PubSeed   [32]byte           // WOTS+ Address Public Seed
	Adrs      [32]byte           // WOTS+ Hash Function Address Scheme
}

// TXHDR represents a Transaction Header structure
type TXHDR struct {
	Options     [4]byte        // Transaction options
	SrcAddr     [ADDR_LEN]byte // Source address
	ChgAddr     [ADDR_LEN]byte // Change address
	SendTotal   [8]byte        // Total amount to send
	ChangeTotal [8]byte        // Total amount to change
	FeeTotal    [8]byte        // Total fee
	BlkToLive   [8]byte        // Block-to-live expiration
}

// TXDAT represents Transaction Data for various TX types
type TXDAT struct {
	Mdst [256]MDST // Multi-Destination array
}

// TXDSA represents Transaction Validation data for various DSA types
type TXDSA struct {
	Wots WOTSVAL // WOTS+ DSA validation data
}

// TXTLR represents a Transaction Trailer structure
type TXTLR struct {
	Nonce [8]byte       // Transaction nonce
	ID    [HASHLEN]byte // Transaction ID
}

// TXENTRY represents a complete transaction entry
type TXENTRY struct {
	Buffer [txBufferSize]byte // Raw transaction buffer
	TxSz   uint64             // Transaction size within buffer

	// Pointers to structures within buffer
	Hdr *TXHDR
	Dat *TXDAT
	Dsa *TXDSA
	Tlr *TXTLR

	// Convenience pointers
	Options     []byte
	SrcAddr     []byte
	ChgAddr     []byte
	SendTotal   []byte
	ChangeTotal []byte
	TxFee       []byte
	TxBtl       []byte
	Mdst        *MDST
	Wots        *WOTSVAL
	TxNonce     []byte
	TxId        []byte
}

// Calculate buffer size needed for TXENTRY
const txBufferSize = unsafe.Sizeof(TXHDR{}) +
	unsafe.Sizeof(TXDAT{}) +
	unsafe.Sizeof(TXDSA{}) +
	unsafe.Sizeof(TXTLR{})

// NewTXENTRY creates and initializes a new TXENTRY
func NewTXENTRY() *TXENTRY {
	tx := &TXENTRY{}

	// Initialize main structure pointers
	tx.Hdr = (*TXHDR)(unsafe.Pointer(&tx.Buffer[0]))
	tx.Dat = (*TXDAT)(unsafe.Pointer(&tx.Buffer[unsafe.Sizeof(TXHDR{})]))
	tx.Dsa = (*TXDSA)(unsafe.Pointer(&tx.Buffer[unsafe.Sizeof(TXHDR{})+unsafe.Sizeof(TXDAT{})]))
	tx.Tlr = (*TXTLR)(unsafe.Pointer(&tx.Buffer[unsafe.Sizeof(TXHDR{})+unsafe.Sizeof(TXDAT{})+unsafe.Sizeof(TXDSA{})]))

	// Initialize convenience pointers
	offset := uintptr(0)
	tx.Options = tx.Buffer[offset : offset+4]
	offset += 4

	tx.SrcAddr = tx.Buffer[offset : offset+ADDR_LEN]
	offset += ADDR_LEN

	tx.ChgAddr = tx.Buffer[offset : offset+ADDR_LEN]
	offset += ADDR_LEN

	tx.SendTotal = tx.Buffer[offset : offset+8]
	offset += 8

	tx.ChangeTotal = tx.Buffer[offset : offset+8]
	offset += 8

	tx.TxFee = tx.Buffer[offset : offset+8]
	offset += 8

	tx.TxBtl = tx.Buffer[offset : offset+8]
	offset += 8

	tx.Mdst = (*MDST)(unsafe.Pointer(&tx.Buffer[unsafe.Sizeof(TXHDR{})]))
	tx.Wots = &tx.Dsa.Wots

	tx.TxNonce = tx.Buffer[len(tx.Buffer)-40 : len(tx.Buffer)-32]
	tx.TxId = tx.Buffer[len(tx.Buffer)-32:]

	return tx
}
