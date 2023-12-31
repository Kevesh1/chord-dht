package main

import (
	"bufio"
	"crypto/rsa"
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
	Backup map[Key]string

	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func CreateNode(args *CreateNodeArgs) {
	node := Node{
		numberSuccessors: args.numberSuccessors,

		Address:     args.Address,
		Bucket:      make(map[Key]string),
		Backup:      make(map[Key]string),
		Successors:  make([]NodeAddress, 4),
		FingerTable: make([]NodeAddress, 160),
		Next:        0,

		timeFixFingers:       args.timeFixFingers,
		timeStabilize:        args.timeStabilize,
		timeCheckPredecessor: args.timeCheckPredecessor,
	}
	createFolders(&node)
	node.generateRSAKey(2048)
	createRing := args.Ring
	if createRing {
		node.create()
	}
	go node.server()
	node.check()
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

func (n *Node) Quit(n1 *struct{}, n2 *struct{}) error {
	successor := n.Successors[0]
	tmp_map := make(map[Key]string)

	for k, v := range n.Bucket {
		tmp_map[k] = v
	}
	ok := call(successor, "Node.Put_all", tmp_map, &struct{}{})
	if !ok {
		fmt.Println("Error moving the keys to the joined node")
	}
	os.Exit(0)
	return nil
}

func (n *Node) AddBackup(args *BackupArgs, None *struct{}) error {
	n.Backup[args.Key] = args.Value
	return nil
}

func (n *Node) GetBucket(none *struct{}, reply *BucketReply) error {
	reply.Bucket = n.Bucket
	return nil
}

func (n *Node) Get(args *GetArgs, reply *GetReply) error {
	fmt.Println("[DEBUG: node.Get()]: args: ", args)
	fmt.Println("[DEBUG: node.Get()]: n.Address: ", n.Address)
	key := args.Key
	if n.Bucket[key] == "true" {
		encryptedBytes := ReadFileBytes("./tmp/" + string(n.Address) + "/" + string(key))
		decryptedBytes, err := n.decrypt(encryptedBytes)
		if err != nil {
			panic(err)
		}
		fmt.Println("[DEBUG: node.Get()]: DECRYPTED output: ")
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
	fmt.Println("Backup:", n.Backup)
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
	encryptedBytes, err := n.encrypt("./samples/" + string(args.FileKey))
	if err != nil {
		fmt.Println("error encrypting file,", args.FileKey)
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

func (n *Node) Put_all(bucket map[Key]string, reply *PutReply) error {
	fmt.Println("[DEBUG node.Put_all()] bucket: ", bucket)
	fmt.Println("[DEBUG node.Put_all()] n.address: ", n.Address)
	for key, value := range bucket {
		n.Bucket[key] = value
		encryptedBytes, err := n.encrypt("./samples/" + string(key))
		if err != nil {
			fmt.Println("[DEBUG node.Put_all()]: error encrypting file,", key)
			panic(err)
		}
		err = os.WriteFile("./tmp/"+string(n.Address)+"/"+string(key), encryptedBytes, 0777)
		if err != nil {
			fmt.Println("error writing file,", key)
			panic(err)
		}
		n.Bucket[key] = "true"
	}
	return nil
}

func (n *Node) Get_all(address NodeAddress, none *struct{}) error {
	insertId := hashString(string(address))
	insertId.Mod(insertId, hashMod)
	nodeId := hashString(string(n.Address))
	nodeId.Mod(nodeId, hashMod)

	tmp_map := make(map[Key]string)
	fmt.Println("[DEBUG node.Get_all()] insterID:", insertId)
	fmt.Println("[DEBUG node.Get_all()] nodeId:", nodeId)
	fmt.Println("[DEBUG node.Get_all()] n.Bucket: ", n.Bucket)

	for k, v := range n.Bucket {
		keyId := hashString(string(k))
		keyId.Mod(keyId, hashMod)
		if !between(insertId, keyId, nodeId, true) {
			fmt.Println("INSIDE IF !BETWEEN")
			tmp_map[k] = v
			delete(n.Bucket, k)
			os.Remove("./tmp/" + string(n.Address) + "/" + string(k))
		}
	}
	ok := call(address, "Node.Put_all", tmp_map, &struct{}{})
	if !ok {
		fmt.Println("Error moving the keys to the joined node")
	}
	return nil
}

// Create chord ring
func (n *Node) create() {
	n.Predecessor = ""
	n.Successors[0] = n.Address
}

func (n *Node) fixFingers() {
	n.Next = n.Next + 1
	if n.Next > fingerTableSize {
		n.Next = 1
	}

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
	if n.Predecessor != "" {
		//kanske jsonrpc istället??
		ok := call(n.Predecessor, "Node.Ping", &HostArgs{}, &reply)
		if !ok {
			fmt.Println("[DEBUG: node.checkPredecessor()] Predecessor is dead")
			n.Predecessor = ""
			for k, v := range n.Backup {
				if v != "" {
					n.Bucket[k] = v
					encryptedBytes, err := n.encrypt("./samples/" + string(k))
					if err != nil {
						fmt.Println("[DEBUG checkPredecessor()]: error encrypting file,", k)
						panic(err)
					}
					err = os.WriteFile("./tmp/"+string(n.Address)+"/"+string(k), encryptedBytes, 0777)
					if err != nil {
						fmt.Println("[DEBUG checkPredecessor()]: error writing file,", k)
						panic(err)
					}
				}
			}
		}
	}
}

func (n *Node) Join(newNode NodeAddress, r *JoinReply) error {

	n.Predecessor = ""

	nodeId := hashString(string(n.Address))
	nodeId.Mod(nodeId, hashMod)

	var reply FindSuccReply
	ok := call(newNode, "Node.Find_successor", nodeId, &reply)
	if ok != true {
		fmt.Println("ERROR")
	}
	n.Successors[0] = reply.Address
	ok = call(n.Address, "Node.Get_all", n.Successors[0], &struct{}{})

	return nil
}

func (n *Node) GetPredecessor(none *struct{}, reply *AddressReply) error {
	reply.Address = n.Predecessor
	return nil
}

// @params: requestID is the current node whos succesor we want to find
// @params: reply is the address of the successor node if one is found
func (n *Node) Find_successor(requestID *big.Int, reply *FindSuccReply) error {
	nodeId := hashString(string(n.Address))
	successorId := hashString(string(n.Successors[0]))
	nodeId.Mod(nodeId, hashMod)
	successorId.Mod(successorId, hashMod)

	if between(nodeId, requestID, successorId, true) {
		reply.Address = n.Successors[0]
		reply.Found = true

	} else {
		nextSuccessor := n.closest_preceding_node(requestID)

		var reply2 FindSuccReply
		ok := call(nextSuccessor, "Node.Find_successor", requestID, &reply2)
		if !ok {
			reply.Found = false
			reply.Address = "Error"
		} else {
			reply.Address = reply2.Address
			reply.Found = true
		}

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

	nodeId := hashString(string(n.Address))
	nodeId.Mod(nodeId, hashMod)
	for i := fingerTableSize; i >= 1; i-- {
		fingerId := hashString(string(n.FingerTable[i]))
		fingerId.Mod(fingerId, hashMod)
		if between(nodeId, fingerId, requestID, false) {
			return n.FingerTable[i]
		}
	}
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
			fmt.Println("Successor is empty, setting successor address to itself")
			n.Successors[0] = n.Address
		} else {
			fmt.Println("Successor is not empty, removing successor AIUFEBIUEIFU")
			for i := 0; i < 4-1; i++ {
				if i == 4-1 {
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

	var bucketReply BucketReply
	ok = call(n.Predecessor, "Node.GetBucket", &struct{}{}, &bucketReply)
	if !ok {
		fmt.Println("Error getting bucket of joined node")
	} else if ok {
		n.Backup = bucketReply.Bucket
	}
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
			time.Sleep(time.Millisecond * 300)
			n.stabilize()
			time.Sleep(time.Millisecond * 300)
			n.fixFingers()
			time.Sleep(time.Millisecond * 300)
			n.checkPredecessor()
		}
	}()
}
