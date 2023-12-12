package main

import (
	"fmt"
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
	addr, err := ClientLookup(key, address)
	if err != nil {
		fmt.Println("[DEBUG node.Get()] Error in ClientLookup(): ", err)
	} else {
		fmt.Println("[DEBUG node.Get()] Found address: ", addr)
	}
	ok := call(addr, "Node.Get", &GetArgs{Key: key}, &GetReply{})
	if !ok {
		fmt.Println("[DEBUG node.Get()] Error in call(): ", err)
	} else {
		// File logic for cypher
	}
}

// Calls Put() inside node.go
func putRPC(address NodeAddress, key Key) {
	addr, err := ClientLookup(key, address)
	if err != nil {
		fmt.Println("[DEBUG node.Put()] Error in ClientLookup(): ", err)
	} else {
		fmt.Println("[DEBUG node.Put()] Found address: ", addr)
	}
	ok := call(addr, "Node.Put", &PutArgs{FileKey: key}, &PutReply{})
	if !ok {
		fmt.Println("[DEBUG node.Put()] Error in call(): ", err)
	} else {
		// File logic for cypher
	}
}

// Calls Delete() inside node.go
func deleteRPC(address NodeAddress, key Key) {
	addr, err := ClientLookup(key, address)
	if err != nil {
		fmt.Println("[DEBUG node.Put()] Error in ClientLookup(): ", err)
	} else {
		fmt.Println("[DEBUG node.Put()] Found address: ", addr)
	}
	ok := call(addr, "Node.Delete", &GetArgs{Key: key}, &GetReply{})
	if !ok {
		fmt.Println("[DEBUG node.Put()] Error in call(): ", err)
	} else {
		// File logic for cypher
	}
}

// Calls Dump() inside node.go
func dumpRPC(address NodeAddress) {
	call(address, "Node.Dump", &GetArgs{}, &GetReply{})
}

func joinRPC(address, joinAddress NodeAddress) {
	call(joinAddress, "Node.Join", address, &JoinReply{})
}

func getPredecessorRPC(address NodeAddress) {
	call(address, "Node.GetPredecessor", struct{}{}, &AddressReply{})

}

func call(address NodeAddress, method string, args interface{}, reply interface{}) bool {
	c, err := rpc.DialHTTP("tcp", string(address))
	if err != nil {
		fmt.Println("Error dialing: ", err)
		return false
	}
	defer c.Close()

	err = c.Call(method, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
