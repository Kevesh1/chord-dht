package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type Key string

type NodeAddress string

var MODULO int = 5

type Node struct {
	Address     NodeAddress
	FingerTable []*Node
	Predecessor *Node
	Successors  []*Node

	Bucket map[Key]string
}

func CreateNode(args *CreateNodeArgs) {
	node := Node{
		Address: args.Address,
		Bucket:  make(map[Key]string),
	}
	node.Bucket["state"] = "abcd"
	go node.server()
	testRPC(args)
	return
}

func (n *Node) Get(args *GetArgs, reply *GetReply) error {
	key := args.Key
	_, ok := n.Bucket[key]
	if !ok {
		return fmt.Errorf("Key not found")
	}
	reply.Value = n.Bucket[key]
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

func (n *Node) stabilize() {
	x := n.Successors[0].Predecessor
	if x == n || x == n.Successors[0] {
		n.Successors[0] = x
	}
	//fmt.Print(successor)

}

func (n *Node) fixFingers() {

}

func (n *Node) checkPredecessor() {

}

func (n *Node) find_successor(id Key) (bool, Node) {
	successor := n.Successors[0]
	if string(n.Address) == string(id) || string(successor.Address) == string(id) {
		return true, Node{Address: successor.Address}
	} else {
		return false, Node{Address: n.closest_preceding_node(id)}
	}
}

func (n *Node) closest_preceding_node(id Key) NodeAddress {
	//TODO:
	//FINGERTABLE LOGIC
	var hash NodeAddress = ""
	return hash
}

func find(id Key, start Node) Node {
	found, nextNode := false, start
	maxSteps := 32
	i := 0
	for found == false && i < maxSteps {
		found, nextNode = nextNode.find_successor(id)
		i++
	}
	if found {
		return nextNode
	} else {
		panic("error")
	}
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
