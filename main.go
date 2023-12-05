package main

import (
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

func (n *Node) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

func (n *Node) server() {
	rpc.Register(n)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":8080")
	//sockname := coordinatorSock()
	//os.Remove(sockname)
	//l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func create() *Node {
	var ip NodeAddress = "192.XXX.XXX"
	n := Node{
		Address: ip,
	}
	n.server()
	return &n
}

func main() {
	n := create()
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
