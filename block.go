package go_mcminterface

import (
	"encoding/binary"
	"fmt"
)

type Block struct {
	Header  BHEADER
	Body    []TXENTRY
	Trailer BTRAILER
}

type BHEADER struct {
	Hdrlen  uint32
	Maddr   [ADDR_TAG_LEN]byte
	Mreward uint64
}

type BTRAILER struct {
	Phash      [HASHLEN]byte
	Bnum       [8]byte
	Mfee       [8]byte
	Tcount     [4]byte
	Time0      [4]byte
	Difficulty [4]byte
	Mroot      [HASHLEN]byte
	Nonce      [HASHLEN]byte
	Stime      [4]byte
	Bhash      [HASHLEN]byte
} // 160

// BHeaderFromBytes - convert bytes to a block header
func bHeaderFromBytes(bytes []byte) BHEADER {
	var header BHEADER

	header.Hdrlen = binary.LittleEndian.Uint32(bytes[0:4])
	if header.Hdrlen != 32 {
		fmt.Println("The block header is corrupted", header.Hdrlen)
		return header
	}
	copy(header.Maddr[:], bytes[4:24])
	header.Mreward = binary.LittleEndian.Uint64(bytes[24:32])

	return header
}

func BBodyFromBytes(bytes []byte) []TXENTRY {
	var body []TXENTRY
	// Iterate through the bytes and create a transaction for each

	padding := 0
	for {
		if padding == len(bytes) {
			break
		} else if padding > len(bytes) {
			fmt.Println("The block was probably corrupted")
			break
		}

		tx, shift := transactionFromBytes(bytes[padding:])
		body = append(body, tx)
		//fmt.Println("Transaction added to the block")
		padding += shift
	}
	//fmt.Println("Block body created", len(body))
	return body
}

// BTrailerFromBytes - convert bytes to a block trailer
func bTrailerFromBytes(bytes []byte) BTRAILER {
	var trailer BTRAILER

	copy(trailer.Phash[:], bytes[0:32])
	copy(trailer.Bnum[:], bytes[32:40])
	copy(trailer.Mfee[:], bytes[40:48])
	copy(trailer.Tcount[:], bytes[48:52])
	copy(trailer.Time0[:], bytes[52:56])
	copy(trailer.Difficulty[:], bytes[56:60])
	copy(trailer.Mroot[:], bytes[60:92])
	copy(trailer.Nonce[:], bytes[92:124])
	copy(trailer.Stime[:], bytes[124:128])
	copy(trailer.Bhash[:], bytes[128:160])

	return trailer
}

// convert bytes to a block
func BlockFromBytes(bytes []byte) Block {
	var block Block

	block.Header = bHeaderFromBytes(bytes)
	block.Body = BBodyFromBytes(bytes[block.Header.Hdrlen : len(bytes)-160])
	block.Trailer = bTrailerFromBytes(bytes[len(bytes)-160:])

	return block
}

// convert a block to bytes
func (bd *Block) GetBytes() []byte {
	var bytes []byte

	bytes = append(bytes, bd.Header.GetBytes()...)
	for _, tx := range bd.Body {
		bytes = append(bytes, tx.Bytes()...)
	}
	bytes = append(bytes, bd.Trailer.GetBytes()...)

	return bytes
}

// convert a block header to bytes
func (bh *BHEADER) GetBytes() []byte {
	var bytes []byte

	// convert to little endian
	hdrlen := make([]byte, 4)
	binary.LittleEndian.PutUint32(hdrlen, bh.Hdrlen)
	bytes = append(bytes, hdrlen...)
	bytes = append(bytes, bh.Maddr[:]...)
	mreward := make([]byte, 8)
	binary.LittleEndian.PutUint64(mreward, bh.Mreward)
	bytes = append(bytes, mreward...)

	return bytes
}

// convert a block trailer to bytes
func (bt *BTRAILER) GetBytes() []byte {
	var bytes []byte

	bytes = append(bytes, bt.Phash[:]...)
	bytes = append(bytes, bt.Bnum[:]...)
	bytes = append(bytes, bt.Mfee[:]...)
	bytes = append(bytes, bt.Tcount[:]...)
	bytes = append(bytes, bt.Time0[:]...)
	bytes = append(bytes, bt.Difficulty[:]...)
	bytes = append(bytes, bt.Mroot[:]...)
	bytes = append(bytes, bt.Nonce[:]...)
	bytes = append(bytes, bt.Stime[:]...)
	bytes = append(bytes, bt.Bhash[:]...)

	return bytes
}
