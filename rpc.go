package main

import (
	"os"
	"strconv"
)

type CreateArgs struct {
}

type CreateReply struct {
}

type CreateNodeArgs struct {
	Address NodeAdress
}

type CreateNodeReply struct {
}

type PortArgs struct {
	Port string
}

type PortReply struct {
}

type HostArgs struct {
	Address NodeAdress
}

type HostReply struct {
}

type JoinArgs struct {
	Address string
}

type JoinReply struct {
}

type PutArgs struct {
	Key   Key
	Value string
}

type PutReply struct {
}

type GetArgs struct {
	Key Key
}

type GetReply struct {
	Value string
}

type DeleteArgs struct {
	Key Key
}

type DeleteReply struct {
}

type DumpArgs struct {
}

type DumpReply struct {
	//TODO dumb all the fields from the node
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
