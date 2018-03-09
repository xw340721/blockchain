package main

import (
	"github.com/boltdb/bolt"
	"log"
	"fmt"
	"encoding/hex"
	"github.com/kardianos/govendor/context"
)

const (
	dbFile       = "blockchain.db"
	blocksBucket = "blocks"
 	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

type (
	BlockChan struct {
		tip []byte
		db *bolt.DB
	}
	BlockchainIterator struct {
		currentHash []byte
		db *bolt.DB
	}
)

func NewBlockChain(address string) *BlockChan {
	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)

	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b==nil{
			fmt.Println("No existing blockchain found. Creating a new one...")
			cbtx := NewCoinbaseTx(address,genesisCoinbaseData)
			genesis := NewGenesisBlock(cbtx)
			b,err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(genesis.Hash,genesis.Serialize())
			if err!=nil{
				log.Panic(err)
			}
			tip = genesis.Hash
		}else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})

	if err!=nil{
		log.Panic(err)
	}

	bc:=BlockChan{
		tip,
		db,
	}
	return &bc
}

func (bc *BlockChan)FindUnspendTransactions(address string)[]Transaction  {
	var unspendTXS []Transaction
	spentTXOs:= make(map[string][]int)
	bci := bc.Iterator()

	for{
		block := bci.Next()

		for _,tx:=range block.Transactions{
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx,out := range tx.Vout{
				if spentTXOs[txID]!=nil{
					for _,spentOut :=range spentTXOs[txID]{
						if spentOut==outIdx{
							continue Outputs
						}
					}
				}
				// 没有地方消耗 就是那些未花费的
				if out.CanBeUnlockedWith(address){
					unspendTXS = append(unspendTXS,*tx)
				}

				if !tx.IsCoinbase(){
					for _,in :=range tx.Vin{
						if in.CanUnlockOutputWith(address){
							inTxID := hex.EncodeToString(in.Txid)
							spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
						}
					}
				}

			}
		}

	}
}

func (bc *BlockChan)FindUTXO(address string)[]Transaction  {
	var UTXOs []Transaction

	unspentTransactions :=bc.FindUnspendTransactions(address)

	for _,tx := range unspentTransactions{
		for _,out :=range tx.Vout{
			if out.CanBeUnlockedWith(address){
				UTXOs = append(UTXOs,out)
			}
		}
	}
	return UTXOs
}

func (bc *BlockChan) AddBlock(data string) {
	var lastHash []byte
	err:=bc.db.View(func(tx *bolt.Tx) error {
		b:= tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})

	if err!=nil{
		log.Panic(err)
	}

	newBlock:= NewBlock(data,lastHash)
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b:=tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash,newBlock.Serialize())
		if err!=nil{
			log.Panic(err)
		}
		err = b.Put([]byte("l"),newBlock.Hash)
		if err!=nil{
			log.Panic(err)
		}
		bc.tip = newBlock.Hash
		return nil
	})
}

func (bc *BlockChan)Iterator()*BlockchainIterator  {
	bci:= &BlockchainIterator{bc.tip,bc.db}
	return bci
}

func (i *BlockchainIterator)Next()*Block{
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b:= tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})

	if err!=nil{
		log.Panic(err)
	}

	// 上一个block的hash
	i.currentHash = block.PrevBlockHash
	return block
}