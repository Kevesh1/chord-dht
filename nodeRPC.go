package main

type NodeRPC struct {
	node *Node
}

func (n *NodeRPC) Create(args *CreateArgs, reply *CreateReply) error {
	n.node = &Node{}
	return nil
}
