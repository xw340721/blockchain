package main

import (
	"fmt"
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
)

const subsidy = 10

type (
	Transaction struct {
		ID   []byte
		Vin  []TXInput
		Vout []TXOutput
	}

	TXInput struct {
		Txid      []byte
		Vout      int
		ScriptSig string
	}

	TXOutput struct {
		Value        int
		ScriptPubKey string
	}
)

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

func NewCoinbaseTx(to,data string)*Transaction  {
	if data==""{
		data = fmt.Sprintf("Reward to %s",to)
	}
	txin:=TXInput{
		[]byte{},
		-1,
		data,
	}

	txout:=TXOutput{
		subsidy,
		to,
	}
	tx := Transaction{
		nil,
		[]TXInput{
			txin,
		},
		[]TXOutput{
			txout,
		},
	}

	tx.SetID()
	return &tx
}

func (tx *Transaction)SetID()  {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err!=nil{
		log.Panic(err)
	}
	hash  = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}