package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Create a scanner to read input from the console
	Arguments := GetCmdArgs()
	fmt.Println(Arguments)
	node := StartChord(Arguments)

	for {
		cliSwitch(node)
	}

}

func cliSwitch(node *Node) {
	scanner := bufio.NewScanner(os.Stdin)
	ipAdd := node.Address

	fmt.Print("Enter command: ")
	// Read a line of input
	scanner.Scan()
	input := scanner.Text()

	// Split the input into command and arguments
	args := strings.Fields(input)
	if len(args) == 0 {
		return
	}

	commandArg := args[0]
	switch commandArg {
	case "help":
		printHelp()

	case "port":
		if len(args) < 2 {
			fmt.Println("Usage: port <port>")
		}
		adress := NodeAddress("127.0.0.1:" + args[1])
		ipAdd = adress
		createNode(&CreateNodeArgs{Address: adress})
		var reply string
		//time.Sleep(time.Second * 1)
		//call(adress, "Node.Ping", &HostArgs{}, &reply)
		fmt.Println("REPLY: ", reply)

	case "create":
		fmt.Println("Creates new ring")

	case "quit":
		quit()

	case "put":
		if len(args) < 3 {
			fmt.Println("Usage: put <key> <value>")
		}
		fmt.Println("Puts key-value pair", args[1], args[2])
		putRPC(ipAdd, Key(args[1]), args[2])

	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: get <key>")
		}
		fmt.Println("Gets value for key", args[1])
		getRPC(ipAdd, Key(args[1]))

	case "clear":
		clearTerminal()

	case "dump":
		dumpRPC(ipAdd)

	default:
		fmt.Println("Unknown command:", commandArg)
		fmt.Println("Use 'help' for more information")
	}
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  help")
	fmt.Println("  port <port>")
	fmt.Println("  create")
	fmt.Println("  quit")
	fmt.Println("  put <key> <value>")
	fmt.Println("  get <key>")
	fmt.Println("  clear")
	fmt.Println("  dump")
}

func clearTerminal() {
	cmd := exec.Command("clear") // Use "clear" for Unix-like systems, "cls" for Windows
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func createRing(args *CreateArgs) *CreateReply {
	fmt.Println("Creates new ring")
	return &CreateReply{}
}

func joinRing(args *JoinArgs) *JoinReply {
	fmt.Println("Joins existing node", args.Address)
	return &JoinReply{}
}

func createNode(args *CreateNodeArgs) {
	fmt.Println("\nCreates new node", args.Address)
	CreateNode(args)
	return
}

func getNode(args *HostArgs) *HostReply {
	fmt.Println("Hosts new node")
	net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.ParseIP(string(args.Address))})
	return &HostReply{}
}

func quit() {
	fmt.Println("\nQuitting program.")
	os.Exit(0)
}
