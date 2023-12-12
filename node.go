package main

import (
	"bufio"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

type Key string

type NodeAddress string

// 2^6 spots --> 64 id's
var MODULO int = 6

var fingerTableSize = 6 // Each finger table i contains the id of (n + 2^i) mod (2^m)th node.
// Use [1, 6] as i and space would be [(n+1)%64, (n+32)%64]

// 2^m
var hashMod *big.Int = new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(MODULO)), nil)

type Node struct {
	//Id *big.Int

	timeFixFingers       int
	timeStabilize        int
	timeCheckPredecessor int

	Address          NodeAddress
	FingerTable      []NodeAddress
	Predecessor      NodeAddress
	Successors       []NodeAddress
	numberSuccessors int

	Next int

	Bucket map[Key]string

	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func (node *Node) generateRSAKey(bits int) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		panic(err)
	}
	node.privateKey = privateKey
	node.publicKey = &privateKey.PublicKey

	// Store private key in Node folder
	priDerText := x509.MarshalPKCS1PrivateKey(privateKey)
	block := pem.Block{
		Type: string(node.Address) + "-private Key",

		Headers: nil,

		Bytes: priDerText,
	}
	node_files_folder := "./tmp/" + node.Address
	privateHandler, err := os.Create(string(node_files_folder) + "/private.pem")
	if err != nil {
		panic(err)
	}
	defer privateHandler.Close()
	pem.Encode(privateHandler, &block)

	// Store public key in Node folder
	pubDerText, err := x509.MarshalPKIXPublicKey(node.publicKey)
	if err != nil {
		panic(err)
	}
	block = pem.Block{
		Type: string(node.Address) + "-public Key",

		Headers: nil,

		Bytes: pubDerText,
	}
	publicHandler, err := os.Create(string(node_files_folder) + "/public.pem")
	if err != nil {
		panic(err)
	}
	defer publicHandler.Close()
	pem.Encode(publicHandler, &block)
}

func CreateNode(args *CreateNodeArgs) {
	node := Node{
		//Id:         hashString(string(args.Address)),
		Address:          args.Address,
		Bucket:           make(map[Key]string),
		Successors:       make([]NodeAddress, 4),
		numberSuccessors: args.numberSuccessors,
		FingerTable:      make([]NodeAddress, 160),
		Next:             0,

		timeFixFingers:       args.timeFixFingers,
		timeStabilize:        args.timeStabilize,
		timeCheckPredecessor: args.timeCheckPredecessor,
	}
	createFolders(&node)
	node.generateRSAKey(2048)
	// node.Id.Mod(node.Id, hashMod)
	createRing := args.Ring
	if createRing {
		node.create()
	}
	node.Bucket["state"] = "abcd"
	go node.server()
	node.check()
	//node.stabilize()
	//testRPC(args)
	return
}

// Helper function to node.CreateNode() for setting up tmp folders for file storage
func createFolders(node *Node) {
	route := "tmp/" + string(node.Address)
	err := os.MkdirAll(route, 0777)
	if err != nil {
		fmt.Println("Error creating tmp folder")
	}
}

// Helper function that copies a file from src to dst
func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func (n *Node) Get(args *GetArgs, reply *GetReply) error {
	key := args.Key
	if n.Bucket[key] == "true" {
		encryptedBytes := ReadFileBytes("./tmp/" + string(n.Address) + "/" + string(key))
		decryptedBytes, err := n.privateKey.Decrypt(nil, encryptedBytes, &rsa.OAEPOptions{Hash: crypto.SHA256})
		if err != nil {
			panic(err)
		}
		fmt.Println("DECRPTED: ")
		fmt.Println(string(decryptedBytes))
	} else {
		return fmt.Errorf("Key not found")
	}
	return nil
}

func (n *Node) Dump(args *GetArgs, reply *GetReply) error {
	//Dumb all attributes of node n
	fmt.Println("Address:", n.Address)
	fmt.Println("FingerTable:", n.FingerTable)
	fmt.Println("Predecessor:", n.Predecessor)
	fmt.Println("Successors:", n.Successors)
	fmt.Println("numberSuccessors:", n.numberSuccessors)
	fmt.Println("Bucket:", n.Bucket)
	fmt.Println("timeFixFingers:", n.timeFixFingers, "ms")
	fmt.Println("timeStabilize:", n.timeStabilize, "ms")
	fmt.Println("timeCheckPredecessor:", n.timeCheckPredecessor, "ms")

	return nil
}

func (n *Node) Delete(args *DeleteArgs, reply *DeleteReply) error {
	key := args.Key
	_, ok := n.Bucket[key]
	if !ok {
		return fmt.Errorf("Key not found")
	}
	delete(n.Bucket, key)
	return nil
}

func (n *Node) Put(args *PutArgs, reply *PutReply) error {
	//copy("./samples/"+string(args.FileKey), "./tmp/"+string(n.Address)+"/"+string(args.FileKey))
	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		n.publicKey,
		ReadFileBytes("./samples/"+string(args.FileKey)),
		nil)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("./tmp/"+string(n.Address)+"/"+string(args.FileKey), encryptedBytes, 0777)
	n.Bucket[args.FileKey] = "true"
	return nil
}

// Helper function to Put() for reading a file into a byte slice
func ReadFileBytes(fileroute string) []byte {
	file, err := os.Open(fileroute)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer file.Close()

	// Get the file size
	stat, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// Read the file into a byte slice
	bs := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(bs)
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return nil
	}
	return bs
}

func (n *Node) Put_all(bucket map[Key]string, reply *PutReply) {
	for key, value := range bucket {
		n.Bucket[key] = value
		encryptedBytes, err := rsa.EncryptOAEP(
			sha256.New(),
			rand.Reader,
			n.publicKey,
			ReadFileBytes("./samples/"+string(key)),
			nil)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile("./tmp/"+string(n.Address)+"/"+string(key), encryptedBytes, 0777)
		n.Bucket[key] = "true"
	}
}

func (n *Node) Get_all(address NodeAddress, none *struct{}) error {
	fmt.Println("GET ALL")
	fmt.Println(address)
	//predecessor := n.Predecessor
	//successorId := hashString(string(n.Successors[0]))
	//successorId.Mod(successorId, hashMod)
	insertId := hashString(string(address))
	insertId.Mod(insertId, hashMod)
	nodeId := hashString(string(n.Address))
	nodeId.Mod(nodeId, hashMod)

	tmp_map := make(map[Key]string)

	//if between(predecessorId, insertId, nodeId, true) {
	for k, v := range n.Bucket {
		keyId := hashString(string(k))
		keyId.Mod(keyId, hashMod)
		if !between(insertId, keyId, nodeId, true) {
			tmp_map[k] = v
			delete(n.Bucket, k)
		}
	}
	//n.Put_all(tmp_map)
	ok := call(address, "Node.Put_all", tmp_map, &struct{}{})
	if !ok {
		fmt.Println("Error moving the keys to the joined node")
	}
	return nil

	//}

}

// Create chord ring
func (n *Node) create() {
	n.Predecessor = ""
	n.Successors[0] = n.Address
	//n.find_successor()
}

func (n *Node) fixFingers() {
	n.Next = n.Next + 1
	if n.Next > fingerTableSize {
		n.Next = 1
	}
	// nodeId := hashString(string(n.Address))
	// nodeId.Mod(nodeId, hashMod)

	requestId := jump(string(n.Address), n.Next)

	nodeId := hashString(string(n.Address))
	nodeId.Mod(nodeId, hashMod)

	var reply FindSuccReply
	ok := call(n.Address, "Node.Find_successor", requestId, &reply)
	if !ok {
		fmt.Println("Error when finding successors in fixFingers")
		return
	}
	if !reply.Found {
		fmt.Println("Could not find successor")
		return
	}
	succesorId := hashString(string(reply.Address))
	succesorId.Mod(succesorId, hashMod)
	n.FingerTable[n.Next] = reply.Address

	for {
		n.Next = n.Next + 1
		if n.Next > fingerTableSize {
			n.Next = 0
			return
		}

		requestId = jump(string(n.Address), n.Next)
		if between(nodeId, requestId, succesorId, false) {
			n.FingerTable[n.Next] = reply.Address
		} else {
			n.Next--
			return
		}
	}

}

func (n *Node) checkPredecessor() {
	var reply string
	if n.Predecessor != n.Address {
		ok := call(n.Predecessor, "Node.Ping", &HostArgs{}, &reply)
		if !ok {
			fmt.Println("[DEBUG: node.checkPredecessor()] Predecessor is dead")
			n.Predecessor = ""
		} else {
		}
	}
}

func (n *Node) Join(newNode NodeAddress, r *JoinReply) error {

	n.Predecessor = ""

	//n.Successors[0] = nodeToJoin
	nodeId := hashString(string(n.Address))
	nodeId.Mod(nodeId, hashMod)

	var reply FindSuccReply
	ok := call(newNode, "Node.Find_successor", nodeId, &reply)
	if ok != true {
		fmt.Println("ERROR")
	}
	n.Successors[0] = reply.Address
	ok = call(n.Successors[0], "Node.Get_all", n.Address, &struct{}{})
	//n.Get_all(n.Successors[0])

	// if between(nodeId, newNodeId, hashString(string(n.Successors[0])), false) || n.Successors[0] == n.Address {
	// 	n.Successors[0] = newNode
	// } else {
	// 	call(n.Successors[0], "Node.Join", newNode, &JoinReply{})
	// }
	return nil
}

func (n *Node) GetPredecessor(none *struct{}, reply *AddressReply) error {
	reply.Address = n.Predecessor
	return nil
}

// @params: requestID is the current node whos succesor we want to find
// @params: reply is the address of the successor node if one is found
func (n *Node) Find_successor(requestID *big.Int, reply *FindSuccReply) error {

	successor := n.Successors[0]
	nodeId := hashString(string(n.Address))
	successorId := hashString(string(successor))
	nodeId.Mod(nodeId, hashMod)
	successorId.Mod(successorId, hashMod)

	//recordHash(nodeId, successorId, requestID)

	if between(nodeId, requestID, successorId, true) {
		reply.Address = successor
		reply.Found = true
		return nil
	} else {

		nextSuccessor := n.closest_preceding_node(requestID)
		fmt.Println("[DEBUG node.FindSuccessor()] Calling for next node: ", successor)
		call(nextSuccessor, "Node.Find_successor", requestID, &FindSuccReply{})
		reply.Address = successor
		reply.Found = false

	}
	return nil
}

func find(requestID *big.Int, start NodeAddress) NodeAddress {
	found, nextNode := false, start
	maxSteps := 32
	i := 0

	for !found && i < maxSteps {
		fmt.Println(i)
		fmt.Println(nextNode)

		result := FindSuccReply{}
		fmt.Println("[DEBUG node.find()] Calling for next node: ", nextNode)
		ok := call(nextNode, "Node.Find_successor", requestID, &result)
		if ok != true {
			fmt.Println("Error in Node.find(): ")
		}
		found = result.Found
		nextNode = result.Address
		i++
	}
	if found {
		fmt.Println("[DEBUG node.find()] Found node A: ", nextNode)
		return nextNode
	}
	fmt.Println("[DEBUG node.find()] Node not found B")
	return "-1"
}

func (n *Node) closest_preceding_node(requestID *big.Int) NodeAddress {
	// skip this loop if you do not have finger tables implemented yet
	//for i = m downto 1
	nodeId := hashString(string(n.Address))
	nodeId.Mod(nodeId, hashMod)
	for i := fingerTableSize; i >= 1; i-- {
		fingerId := hashString(string(n.FingerTable[i]))
		fingerId.Mod(fingerId, hashMod)
		if between(nodeId, fingerId, requestID, false) {
			return n.FingerTable[i]
		}
	}
	// if (finger[i] ∈ (n,id])
	// 	return finger[i];
	return n.Successors[0]
}

func (n *Node) server() {
	rpc.Register(n)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", string(n.Address))
	fmt.Println("Local node listening on ", n.Address)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)
}

func (n *Node) GetSuccessors(none *struct{}, reply *SuccessorsListReply) error {
	reply.Successors = n.Successors
	return nil
}

func (n *Node) stabilize() {
	successor := n.Successors[0]
	var successorsReply SuccessorsListReply
	ok := call(successor, "Node.GetSuccessors", &struct{}{}, &successorsReply)
	successors := successorsReply.Successors
	if ok {
		for i := 0; i < n.numberSuccessors-2; i++ {
			n.Successors[i+1] = successors[i]

		}
	} else {
		fmt.Println("GetSuccessors failed")
		if successor == "" {
			n.Successors[0] = n.Address
		} else {
			for i := 0; i < n.numberSuccessors-1; i++ {
				if i == n.numberSuccessors-1 {
					n.Successors[i] = ""
				} else {
					n.Successors[i] = n.Successors[i+1]
				}
			}
		}
	}

	var reply AddressReply
	call(successor, "Node.GetPredecessor", &struct{}{}, &reply)
	predecessor := reply.Address
	nodeId := hashString(string(n.Address))
	nodeId.Mod(nodeId, hashMod)
	predecessorID := hashString(string(predecessor))
	predecessorID.Mod(predecessorID, hashMod)
	successorID := hashString(string(successor))
	successorID.Mod(successorID, hashMod)
	if between(nodeId, predecessorID, successorID, false) {
		n.Successors[0] = predecessor
	}
	call(successor, "Node.Notify", n.Address, &struct{}{})
	//fmt.Print(successor)
}

func (n *Node) Notify(address NodeAddress, none *struct{}) error {
	predecessorID := hashString(string(n.Predecessor))
	predecessorID.Mod(predecessorID, hashMod)

	addressID := hashString(string(address))
	addressID.Mod(addressID, hashMod)

	nodeId := hashString(string(n.Address))
	nodeId.Mod(nodeId, hashMod)

	if n.Predecessor == "" || between(predecessorID, addressID, nodeId, false) {
		n.Predecessor = address
	}
	return nil
}
func (n *Node) Ping(args *HostArgs, reply *string) error {
	*reply = "Ping received"
	return nil
}

func (n *Node) check() {
	go func() {
		for {
			time.Sleep(time.Millisecond * time.Duration(n.timeStabilize))
			n.stabilize()
			time.Sleep(time.Millisecond * time.Duration(n.timeFixFingers))
			//	n.fixFingers()
			time.Sleep(time.Millisecond * time.Duration(n.timeCheckPredecessor))
			n.checkPredecessor()
		}
	}()
}
