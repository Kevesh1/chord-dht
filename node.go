package main

import (
	"fmt"
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

	Address     NodeAddress
	FingerTable []NodeAddress
	Predecessor NodeAddress
	Successors  []NodeAddress

	Next int

	Bucket map[Key]string
}

func CreateNode(args *CreateNodeArgs) {
	node := Node{
		//Id:         hashString(string(args.Address)),
		Address:     args.Address,
		Bucket:      make(map[Key]string),
		Successors:  make([]NodeAddress, 4),
		FingerTable: make([]NodeAddress, 160),
		Next:        0,
	}
	// node.Id.Mod(node.Id, hashMod)
	createRing := args.Ring
	if createRing {
		node.create()
		fmt.Println("INSIDE")
	}
	node.Bucket["state"] = "abcd"
	go node.server()
	node.check()
	//node.stabilize()
	//testRPC(args)
	return
}

func (n *Node) Get(args *GetArgs, reply *GetReply) error {
	key := args.Key
	_, ok := n.Bucket[key]
	if !ok {
		return fmt.Errorf("Key not found")
	}
	fmt.Println(n.Bucket[key])
	return nil
}

func (n *Node) Dump(args *GetArgs, reply *GetReply) error {
	//Dumb all attributes of node n
	fmt.Println("Address:", n.Address)
	fmt.Println("FingerTable:", n.FingerTable)
	fmt.Println("Predecessor:", n.Predecessor)
	fmt.Println("Successors:", n.Successors)
	fmt.Println("Bucket:", n.Bucket)
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
	key := args.Key
	// _, ok := n.Bucket[key]
	// if !ok {
	// 	return fmt.Errorf("Key not found")
	// }
	n.Bucket[key] = args.Value
	return nil
}

func (n *Node) Put_all(bucket map[Key]string, reply *PutReply) error {
	fmt.Println("PUT ALL")
	for key, value := range bucket {
		n.Bucket[key] = value
	}
	return nil
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
	//i := 0
	//fmt.Println(i)
	//i++
	successor := n.Successors[0]
	var successorsReply SuccessorsListReply
	ok := call(successor, "Node.GetSuccessors", &struct{}{}, &successorsReply)
	successors := successorsReply.Successors
	if ok {
		for i := 0; i < 4-2; i++ {
			//fmt.Println(i)
			// fmt.Println(len(n.Successors))
			// fmt.Println(len(successors))
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

	//fmt.Println("AAAA")

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
	//fmt.Println("INSIDE")
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
