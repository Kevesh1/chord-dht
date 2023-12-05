package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Key string

type NodeAdress string

type Node struct {
	ID          Key
	Address     NodeAdress
	Fingertable []NodeAdress
	Successor   []NodeAdress
	Predecessor NodeAdress

	Bucket map[Key]string
}

func (n *Node) server() {
	nrpc := &NodeRPC{n}
	rpc.Register(nrpc)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", string(n.Address))
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func (n *Node) Join(address NodeAdress) {
	//TODO
}

func (n *Node) Quit() {
	//TODO
}

func (n *Node) Put(key Key, value string) {
	//TODO
}

func (n *Node) Get(args *GetArgs) *GetReply {
	//TODO
	return &GetReply{Value: "TODO"}
}

func (n *Node) Delete(key Key) {
	//TODO
}

func call(address string, method string, request interface{}, reply interface{}) error {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	err = client.Call(method, request, reply)
	if err != nil {
		log.Fatal("arith error:", err)
	}
	return err
}

func (n *Node) getLocalAddress() string {
	conn, err := net.Dial("udp", "0.0.0.0:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
