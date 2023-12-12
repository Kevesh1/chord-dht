package main

func (n *Node) Create(args *CreateArgs, reply *CreateReply) error {
	return nil
}

type CreateArgs struct {
}

type CreateReply struct {
}

type CreateNodeArgs struct {
	Address              NodeAddress
	Ring                 bool
	timeFixFingers       int
	timeStabilize        int
	timeCheckPredecessor int
	numberSuccessors     int
}

type CreateNodeReply struct {
}

type PortArgs struct {
	Port string
}

type PortReply struct {
}

type HostArgs struct {
	Address NodeAddress
}

type HostReply struct {
}

type JoinArgs struct {
	Address NodeAddress
}

type JoinReply struct {
}

type PutArgs struct {
	FileKey Key
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

type FindSuccReply struct {
	Address NodeAddress
	Found   bool
}

type AddressReply struct {
	Address NodeAddress
}

type SuccessorsListReply struct {
	Successors []NodeAddress
}
