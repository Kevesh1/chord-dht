package main

import "fmt"

func main() {

	//Read flags from program start and check if valid
	args := ReadFlagsCmd()
	CheckArgsValid(args)
	fmt.Println(args)

	ipAdd := NodeAddress(string(args.Address) + ":" + fmt.Sprint(args.Port))

	if args.JoinAddress != "Unspecified" {
		createNode(&CreateNodeArgs{Address: ipAdd, Ring: true})
		fullAddress := NodeAddress(string(args.JoinAddress) + ":" + fmt.Sprint(args.JoinPort))
		joinRPC(ipAdd, fullAddress)
	} else {
		createNode(&CreateNodeArgs{
			Address:              ipAdd,
			Ring:                 false,
			timeFixFingers:       args.tff,
			timeStabilize:        args.ts,
			timeCheckPredecessor: args.tcp,
			numberSuccessors:     args.Successors,
		})

	}

	// Command line interface, handles all commands listed in: https://computing.utahtech.edu/cs/3410/asst_chord.html
	cli(ipAdd)
}
