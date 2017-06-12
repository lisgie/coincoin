package bc

import (
	"crypto/ecdsa"
	"testing"
	"os"
	"math/big"
	"crypto/elliptic"
	"golang.org/x/crypto/sha3"
	"crypto/rand"

	"storage"
	"io/ioutil"
	"log"
)

var accA, accB, minerAcc *Account
var PrivKeyA, PrivKeyB ecdsa.PrivateKey
var PubKeyA, PubKeyB ecdsa.PublicKey
var RootPrivKey ecdsa.PrivateKey


func addTestingAccounts() {

	accA,accB,minerAcc = new(Account),new(Account),new(Account)

	puba1,_ := new(big.Int).SetString(pubA1,16)
	puba2,_ := new(big.Int).SetString(pubA2,16)
	priva,_ := new(big.Int).SetString(privA,16)
	PubKeyA = ecdsa.PublicKey{
		elliptic.P256(),
		puba1,
		puba2,
	}
	PrivKeyA = ecdsa.PrivateKey{
		PubKeyA,
		priva,
	}

	pubb1,_ := new(big.Int).SetString(pubB1,16)
	pubb2,_ := new(big.Int).SetString(pubB2,16)
	privb,_ := new(big.Int).SetString(privB,16)
	PubKeyB = ecdsa.PublicKey{
		elliptic.P256(),
		pubb1,
		pubb2,
	}
	PrivKeyB = ecdsa.PrivateKey{
		PubKeyB,
		privb,
	}

	accA.Balance = 123232345678
	copy(accA.Address[0:32], PrivKeyA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], PrivKeyA.PublicKey.Y.Bytes())
	accA.Hash = sha3.Sum256(accA.Address[:])

	//This one is just for testing purposes
	accB.Balance = 823237654321
	copy(accB.Address[0:32], PrivKeyB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyB.PublicKey.Y.Bytes())
	accB.Hash = sha3.Sum256(accB.Address[:])

	//just to bootstrap
	var shortHashA [8]byte
	var shortHashB [8]byte
	copy(shortHashA[:], accA.Hash[0:8])
	copy(shortHashB[:], accB.Hash[0:8])

	State[shortHashA] = append(State[shortHashA],accA)
	State[shortHashB] = append(State[shortHashB],accB)

	MinerPrivKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var pubKey [64]byte
	var shortMiner [8]byte
	copy(pubKey[:32],MinerPrivKey.X.Bytes())
	copy(pubKey[32:],MinerPrivKey.Y.Bytes())
	MinerHash = serializeHashContent(pubKey[:])
	copy(shortMiner[:],MinerHash[0:8])
	minerAcc.Hash = MinerHash
	minerAcc.Address = pubKey
	State[shortMiner] = append(State[shortMiner],minerAcc)

}

func addRootAccounts() {

	var pubKey [64]byte

	pub1,_ := new(big.Int).SetString(RootPub1,16)
	pub2,_ := new(big.Int).SetString(RootPub2,16)
	priv,_ := new(big.Int).SetString(RootPriv,16)
	PubKeyA = ecdsa.PublicKey{
		elliptic.P256(),
		pub1,
		pub2,
	}
	RootPrivKey = ecdsa.PrivateKey{
		PubKeyA,
		priv,
	}

	copy(pubKey[32-len(pub1.Bytes()):32],pub1.Bytes())
	copy(pubKey[64-len(pub2.Bytes()):],pub2.Bytes())

	rootHash := serializeHashContent(pubKey[:])

	var shortRootHash [8]byte
	copy(shortRootHash[:], rootHash[0:8])
	rootAcc := Account{Hash:rootHash, Address:pubKey}
	State[shortRootHash] = append(State[shortRootHash], &rootAcc)
	RootKeys[rootHash] = &rootAcc
}

func TestMain(m *testing.M) {

	//initialize states
	State = make(map[[8]byte][]*Account)
	RootKeys = make(map[[32]byte]*Account)

	storage.Init()

	//genesis block
	genesis := newBlock()
	writeBlock(genesis)
	collectStatistics(genesis)

	//setting a new random seed
	addTestingAccounts()
	addRootAccounts()
	//we don't want logging msgs when testing, designated messages
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}