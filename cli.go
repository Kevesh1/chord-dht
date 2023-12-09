package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"regexp"
)

type Arguments struct {
	// Read command line arguments
	Address     NodeAddress // Current node address
	Port        int         // Current node port
	JoinAddress NodeAddress // Joining node address
	JoinPort    int         // Joining node port
	Successors  int
	ClientName  string
}

func GetCmdArgs() Arguments {
	// Read command line arguments
	var a string  // Current node address
	var p int     // Current node port
	var ja string // Joining node address
	var jp int    // Joining node port
	var r int     // The number of successors to maintain.
	var i string  // Client name

	// Parse command line arguments
	flag.StringVar(&a, "a", "localhost", "Current node address")
	flag.IntVar(&p, "p", 8000, "Current node port")
	flag.StringVar(&ja, "ja", "Unspecified", "Joining node address")
	flag.IntVar(&jp, "jp", 8000, "Joining node port")
	flag.IntVar(&r, "r", 3, "The number of successors to maintain.")
	flag.StringVar(&i, "i", "Default", "Client ID/Name")
	flag.Parse()

	// Return command line arguments
	return Arguments{
		Address:     NodeAddress(a),
		Port:        p,
		JoinAddress: NodeAddress(ja),
		JoinPort:    jp,
		Successors:  r,
		ClientName:  i,
	}
}

func CheckArgsValid(args Arguments) int {
	if !checkIPValid(string(args.Address)) {
		fmt.Println("IP address is invalid")
		return -1
	}
	// Check if port is valid
	if !checkPortsValid(args.Port) {
		fmt.Println("Port number is invalid")
		return -1
	}

	// Check if number of successors is valid
	if !checkSuccessorsValid(args.Successors) {
		fmt.Println("Successors number is invalid")
		return -1
	}

	if !clientNameValid(args.ClientName) {
		fmt.Println("Client name is invalid")
		return -1
	}

	// Check if joining address and port is valid or not
	if checkJoinArgsValid(args) {
		fmt.Println("Joining address is invalid")
		return -1
	}
	return 0
}

// <--- START Checks for Valid Arguments --->
func checkIPValid(ip string) bool {
	if ip != "localhost" {
		return false
	}
	return true
}

func checkPortsValid(port int) bool {
	if port < 1024 || port > 65535 {
		return false
	}
	return true
}

func checkSuccessorsValid(successors int) bool {
	if successors < 1 || successors > 32 {
		return false
	}
	return true
}

func clientNameValid(clientName string) bool {
	matched, err := regexp.MatchString("[0-9a-fA-F]*", clientName)
	if err != nil || !matched {
		return false
	}
	return true
}

func checkJoinArgsValid(args Arguments) bool {
	if args.JoinAddress == "Unspecified" {
		if net.ParseIP(string(args.Address)) != nil || args.Address == "localhost" {
			if checkPortsValid(args.Port) {
				return false
			}
		}
	}
	return true
}

// <--- END Checks for Valid Arguments --->

func StartChord(args Arguments) *Node {
	// Check if the command line arguments are valid
	valid := CheckArgsValid(args)
	var node *Node
	if valid == -1 {
		fmt.Println("Invalid command line arguments")
		os.Exit(1)
	} else {
		fmt.Println("Valid command line arguments")
		// Create new Node
		node := CreateNode(&CreateNodeArgs{Address: args.Address})

		rpc.Register(node)

		go node.server()

		if valid == 0 {
			// Join exsiting chord
			RemoteAddr := fmt.Sprintf("%s:%d", args.JoinAddress, args.JoinPort)

			// Connect to the remote node
			fmt.Println("Connecting to the remote node..." + RemoteAddr)
			err := joinRPC(args.Address, args.JoinAddress)
			if err != nil {
				fmt.Println("Join RPC call failed")
				os.Exit(1)
			} else {
				fmt.Println("Join RPC call success")
			}
		} else if valid == 1 {
			// Create new chord
			node.Predecessor = ""
			// All successors are itself when create a new Chord ring
			for i := 0; i < len(node.Successors); i++ {
				node.Successors[i] = node.Address
			}
			// Combine address and port, convert port to string
		}

		node.check()
	}
	return node
}
