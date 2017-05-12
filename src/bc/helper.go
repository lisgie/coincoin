package bc

import (
	"bytes"
	"encoding/binary"
	"golang.org/x/crypto/sha3"
	"encoding/gob"
	"math/big"
)

func serializeHashContent(data interface{}) (hash [32]byte) {
	// Create a struct and write it.
	var buf bytes.Buffer

	binary.Write(&buf,binary.LittleEndian, data)

	return sha3.Sum256(buf.Bytes())
}

func validateProofOfWork(diff uint8, hash [32]byte) bool {
	var byteNr uint8
	for byteNr = 0; byteNr < (uint8)(diff/8); byteNr++ {
		if hash[byteNr] != 0 {
			return false
		}
	}
	if diff%8 != 0 && hash[byteNr+1] >= 1<<(8-diff%8) {
		return false
	}
	return true
}

func proofOfWork(diff uint8, merkleRoot [32]byte) *big.Int {

	var tmp [32]byte
	var byteNr uint8
	var abort bool
	//big int needed because int64 overflows if nonce too large
	oneIncr := big.NewInt(1)
	cnt := big.NewInt(0)

	for ;; cnt.Add(cnt,oneIncr) {
		abort = false

		tmp = sha3.Sum256(append(cnt.Bytes(),merkleRoot[:]...))
		for byteNr = 0; byteNr < (uint8)(diff/8); byteNr++ {
			if tmp[byteNr] != 0 {
				abort = true
				break
			}
		}
		if abort {
			continue
		}

		if diff%8 != 0 && tmp[byteNr+1] >= 1<<(8-diff%8) {
			continue
		}
		break
	}

	return cnt
}

//gob.Register for interface implementations
func EncodeForSend(data interface{}) []byte {

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(data)
	return buf.Bytes()
}


/*func DecodeForReceive(payload []byte) Block {

	var decoded
	var buf bytes.Buffer

	dec := gob.NewDecoder(&buf)
	dec.Decode(decoded)
	return Block(buf.Bytes())
}*/