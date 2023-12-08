package main

import (
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
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

	Bucket map[Key]string
}

func CreateNode(args *CreateNodeArgs) {
	node := Node{
		//Id:         hashString(string(args.Address)),
		Address:    args.Address,
		Bucket:     make(map[Key]string),
		Successors: make([]NodeAddress, 1),
	}
	// node.Id.Mod(node.Id, hashMod)
	node.create()
	node.Bucket["state"] = "abcd"
	go node.server()
	//testRPC(args)
	return
}

func (n *Node) Get(args *GetArgs, reply *GetReply) error {
	key := args.Key
	_, ok := n.Bucket[key]
	if !ok {
		return fmt.Errorf("Key not found")
	}
	fmt.Println(n.Bucket)
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
	_, ok := n.Bucket[key]
	if !ok {
		return fmt.Errorf("Key not found")
	}
	n.Bucket[key] = args.Value
	return nil
}

// Create chord ring
func (n *Node) create() {
	n.Predecessor = ""
	n.Successors[0] = n.Address
	fmt.Println(n.Successors)
	fmt.Println(len(n.Successors))
	fmt.Println(n.Successors[0])
	//n.find_successor()
}

func (n *Node) fixFingers() {

}

func (n *Node) checkPredecessor() {

}

func (n *Node) Join(nodeToJoin NodeAddress, r *JoinReply) error {

	n.Predecessor = ""

	//n.Successors[0] = nodeToJoin
	NodeId := hashString(string(n.Address))
	NodeId.Mod(NodeId, hashMod)

	var reply FindSuccReply
	ok := call(nodeToJoin, "Node.Find_successor", n.Address, &reply)
	if ok != true {
		fmt.Println("ERROR")
	}
	n.Successors[0] = reply.Address
	return nil
}

func (n *Node) GetPredecessor(none *struct{}, reply *AddressReply) error {
	reply.Address = n.Predecessor
	return nil
}

func (n *Node) Find_successor(nodeToBeJoinedOn NodeAddress, reply *FindSuccReply) error {
	successor := n.Successors[0]
	fmt.Println("FIRST INDEX: ", successor)
	fmt.Println("SECOND INDEX: ", n.Successors[1])
	nodeId := hashString(string(n.Address))
	nodeId.Mod(nodeId, hashMod)
	nodeToBeJoinedOnID := hashString(string(nodeToBeJoinedOn))
	nodeToBeJoinedOnID.Mod(nodeToBeJoinedOnID, hashMod)
	successorID := hashString(string(successor))
	successorID.Mod(successorID, hashMod)

	if between(nodeId, nodeToBeJoinedOnID, successorID, false) {
		reply.Address = successor
		reply.Found = true
		return nil
	} else {
		fmt.Println("[DEBUG node.FindSuccessor()] Calling for next node: ", successor)
		call(successor, "Node.Find_successor", nodeToBeJoinedOn, &FindSuccReply{})
		reply.Address = "AAAAA"
		reply.Found = false
		//return n.find_successor(id)
		//return false, Node{Address: n.closest_preceding_node(id)}
	}
	return nil
}

func find(start NodeAddress) NodeAddress {
	found, nextNode := false, start
	maxSteps := 32
	i := 0

	for !found && i < maxSteps {
		nodeId := hashString(string(nextNode))
		nodeId.Mod(nodeId, hashMod)
		result := FindSuccReply{}
		fmt.Println("[DEBUG node.find()] Calling for next node: ", nextNode)
		ok := call(nextNode, "Node.Find_successor", nodeId, &result)
		if ok != true {
			fmt.Println("Error in Node.find(): ")
		}
		found = result.Found
		nextNode = result.Address
		i++
	}
	if found {
		return nextNode
	}
	return "-1"
}

func (n *Node) closest_preceding_node(id Key) NodeAddress {
	//TODO:
	//FINGERTABLE LOGIC
	var hash NodeAddress = ""
	return hash
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

func (n *Node) stabilize() {
	successor := n.Successors[0]
	var reply AddressReply
	call(successor, "Node.GetPredecessor", struct{}{}, &reply)
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
	call(successor, "node.Notify", n.Address, struct{}{})
	//fmt.Print(successor)

}

func (n *Node) Notify(address NodeAddress) {
	predecessorID := hashString(string(n.Predecessor))
	predecessorID.Mod(predecessorID, hashMod)

	addressID := hashString(string(address))
	addressID.Mod(addressID, hashMod)

	nodeId := hashString(string(n.Address))
	nodeId.Mod(nodeId, hashMod)

	if n.Predecessor == "" || between(predecessorID, addressID, nodeId, false) {
		n.Predecessor = address
	}
}

func (n *Node) Ping(args *HostArgs, reply *string) error {
	fmt.Println("INSIDE")
	*reply = "Ping received"
	return nil
}

func (n *Node) check() {
	go func() {
		for {
			n.stabilize()
			time.Sleep(time.Second * (1 / 3))
			n.fixFingers()
			time.Sleep(time.Second * (1 / 3))
			n.checkPredecessor()
			time.Sleep(time.Second * (1 / 3))
		}
	}()
}
