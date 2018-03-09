package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
	"time"
)

const targetBit = 24

var (
	maxNonce = math.MaxInt64
)

type (
	ProofOfWork struct {
		block  *Block
		target *big.Int
	}
)

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	// 每次运行都是同一target目标
	target.Lsh(target, uint(256-targetBit))

	pow := &ProofOfWork{
		b,
		target,
	}
	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBit)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hash [32]byte
	var hashInt big.Int
	now := time.Now()
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	spendTime := time.Now().Sub(now).Minutes()
	fmt.Printf("\r\nspend minutes %f", spendTime)
	fmt.Print("\n\n")
	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
}

func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}

func main() {
	bc := NewBlockChain()
	defer bc.db.Close()
	cli := CLI{bc}
	cli.Run()

}
