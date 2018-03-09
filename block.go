package main

import (
	"bytes"
	"encoding/gob"
	"time"
	"log"
	"crypto/sha256"
)

type Block struct {
	Timestamp     int64
	Transactions          []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		time.Now().Unix(),
		transactions,
		prevBlockHash,
		[]byte{},
		0,
	}
	// 每次创建 即运算
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}


func (b *Block)Serialize()[]byte{
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)

	if err!=nil{
		log.Panic(err)
	}

	return result.Bytes()
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func DeserializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	decoder.Decode(&block)

	return &block
}

func (b *Block)HashTransactions()[]byte  {
	var txHashes [][]byte
	var txHash [32]byte

	for _,tx:=range b.Transactions{
		txHashes = append(txHashes,tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes,[]byte{}))
	return txHash[:]
}