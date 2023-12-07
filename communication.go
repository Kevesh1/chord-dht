package main

import (
	"fmt"
	"log"
	"net/rpc"
)

func testRPC(args *CreateNodeArgs) {
	var reply string
	//time.Sleep(time.Second * 5)
	fmt.Println("BUUUUUZZZZ")
	ok := call(args.Address, "Node.Ping", &HostArgs{}, &reply)
	if !ok {
		fmt.Println("Error calling for task")
	}
	fmt.Println("AFTER")
	fmt.Println("REPLY: ", reply)
}

// Calls Get() inside node.go
// These 3 methods work as the middleman atm
func getRPC(address NodeAddress, key Key) {
	call(address, "Node.Get", &GetArgs{Key: key}, &GetReply{})
}

// Calls Put() inside node.go
func putRPC(address NodeAddress, key Key, value string) {
	call(address, "Node.Put", &PutArgs{Key: key, Value: value}, &PutReply{})

}

// Calls Delete() inside node.go
func deleteRPC(address NodeAddress, key Key) {
	call(address, "Node.Delete", &GetArgs{Key: key}, &GetReply{})
}

// Calls Dump() inside node.go
func dumpRPC(address NodeAddress) {
	call(address, "Node.Dump", &GetArgs{}, &GetReply{})
}

func joinRPC(address, joinAddress NodeAddress) {
	call(address, "Node.Join", joinAddress, &JoinReply{})
}

func getPredecessorRPC(address NodeAddress) {
	call(address, "Node.GetPredecessor", struct{}{}, &AddressReply{})

}

func call(address NodeAddress, method string, args interface{}, reply interface{}) bool {
	c, err := rpc.DialHTTP("tcp", string(address))
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(method, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
