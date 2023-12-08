package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"math/big"
	"os"
)

// Records values and states when hashing node vals
// Values are stored in hash_records.txt, 'H' stands for hashed
// @params:HnodeID, HrequestID, HsucessorID
func recordHash(
	HnodeId *big.Int, HrequestId *big.Int, HsucessorId *big.Int,
) error {

	file, err := os.OpenFile("hash_records.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString("\n <-- NEW RECORD --> \n")
	//Write nodeID before and after hashed
	file.WriteString("tHNodeID: " + HnodeId.String() + "\n")
	//Write requestID before and after hashed
	file.WriteString("HRequestID: " + HrequestId.String() + "\n")
	//Write sucessorID before and after hashed
	file.WriteString("HSucessorID: " + HsucessorId.String() + "\n")
	fmt.Println("Hashes recorded in hash_records.txt")
	return nil
}

func hashString(elt string) *big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(elt))
	return new(big.Int).SetBytes(hasher.Sum(nil))
}

const keySize = sha1.Size * 8

func jump(address string, fingerentry int) *big.Int {
	n := hashString(address)
	fingerentryminus1 := big.NewInt(int64(fingerentry) - 1)
	two := big.NewInt(2)
	jump := new(big.Int).Exp(two, fingerentryminus1, nil)
	sum := new(big.Int).Add(n, jump)

	return new(big.Int).Mod(sum, hashMod)
}

func between(start, elt, end *big.Int, inclusive bool) bool {
	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}
}

func ClientLookup(key Key, nodeAdr NodeAddress) (NodeAddress, error) {
	keyId := hashString(string(key))
	keyId.Mod(keyId, hashMod)
	addr := find(keyId, nodeAdr)

	if addr == "-1" {
		return "", errors.New("Key not found")
	}
	return NodeAddress(addr), nil
}
