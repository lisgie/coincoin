package bc

import (
	"log"
	"golang.org/x/crypto/sha3"
	"bytes"
	"errors"
	"encoding/binary"
)

func getAccountFromHash(hash [32]byte) (*Account) {

	var fixedHash [8]byte
	copy(fixedHash[:],hash[0:8])
	for _,acc := range State[fixedHash] {
		if bytes.Compare(acc.Hash[:],hash[:]) == 0 {
			return acc
		}
	}
	return nil
}

func isRootKey(hash [32]byte) (bool) {

	for _,rootAcc := range RootKeys {
		if rootAcc.Hash == hash {
			return true
		}
	}
	return false
}

//possibility of state change
//1) exchange funds from tx
//2) revert funds from previous tx
//3) this doesn't need a rollback, because digitally signed
//https://golang.org/doc/faq#stack_or_heap
func accStateChange(acctx *accTx) {

	var fixedHash [8]byte
	addressHash := sha3.Sum256(acctx.PubKey[:])
	acc := getAccountFromHash(addressHash)
	if acc != nil {
		log.Printf("Address already exists in the state: %x\n", addressHash[0:4])
		return
	}
	copy(fixedHash[:],addressHash[0:8])
	newAcc := Account{Hash:addressHash,Address:acctx.PubKey}
	State[fixedHash] = append(State[fixedHash],&newAcc)
}

func fundsStateChange(tx *fundsTx) error {

	accSender, accReceiver := getAccountFromHash(tx.fromHash), getAccountFromHash(tx.toHash)

	if accSender == nil {
		log.Printf("Sender does not exist in the State: %x\n", tx.fromHash[0:8])
		return errors.New("Sender does not exist in the State.")
	}

	if accReceiver == nil {
		log.Printf("Receiver does not exist in the State: %x\n", tx.toHash[0:8])
		return errors.New("Receiver does not exist in the State.")
	}

	//also check for txCnt!
	var cntBuf [4]byte
	copy(cntBuf[1:],tx.TxCnt[:])
	txCnt := binary.BigEndian.Uint32(cntBuf[:])
	if txCnt != accSender.TxCnt {
		log.Printf("Sender txCnt does not match: %v (tx.txCnt) vs. %v (state txCnt)\n", txCnt, accSender.TxCnt)
		return errors.New("Sender does not have enough funds for the transaction.")
	}

	amount := binary.BigEndian.Uint64(tx.Amount[:])
	fee := binary.BigEndian.Uint64(tx.Fee[:])
	if (amount + fee) > accSender.Balance {
		log.Printf("Sender does not have enough balance: %x\n", accSender.Balance)
		return errors.New("Sender does not have enough funds for the transaction.")
	}

	//we're manipulating pointer, no need to write back
	accSender.TxCnt += 1
	accSender.Balance -= amount

	accReceiver.Balance += amount

	return nil
}

func collectTxFees(fundsTx []*fundsTx, accTx []*accTx, minerHash [32]byte) {
	miner := getAccountFromHash(minerHash)

	//subtract fees from sender (check if that is allowed has already been done in the block validation)
	for _,tx := range fundsTx {
		fee := binary.BigEndian.Uint64(tx.Fee[:])
		miner.Balance += fee

		senderAcc := getAccountFromHash(tx.fromHash)
		senderAcc.Balance -= fee
	}

	for _,tx := range accTx {
		//money gets created from thin air
		//no need to subtract money from root key
		fee := binary.BigEndian.Uint64(tx.Fee[:])
		miner.Balance += fee
	}
}

func collectBlockReward(reward uint64, minerHash [32]byte) {
	miner := getAccountFromHash(minerHash)
	miner.Balance += reward
}

func fundsStateRollback(txSlice []*fundsTx, index int) {

	for cnt := index; index >= 0; index-- {
		tx := txSlice[cnt]

		accSender, accReceiver := getAccountFromHash(tx.fromHash), getAccountFromHash(tx.toHash)

		amount := binary.BigEndian.Uint64(tx.Amount[:])
		accSender.TxCnt -= 1
		accSender.Balance += amount

		accReceiver.Balance -= amount
	}
}

func PrintState() {
	log.Println("State updated: ")
	for key,val := range State {
		for _,acc := range val {
			log.Printf("%x: %v\n", key[0:4], acc)
		}
	}
}
