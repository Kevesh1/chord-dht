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
	fmt.Println("Welcome to the CLI program. Type 'help' for commands.")

	// Create a scanner to read input from the console
	scanner := bufio.NewScanner(os.Stdin)

	// Run the program continuously until the "quit" command is entered
	for {
		fmt.Print("Enter command: ")
		// Read a line of input
		scanner.Scan()
		input := scanner.Text()

		// Split the input into command and arguments
		args := strings.Fields(input)
		if len(args) == 0 {
			continue
		}

		// Extract the command and handle it
		commandArg := args[0]
		switch commandArg {
		case "help":
			printHelp()

		case "port":
			if len(args) < 2 {
				fmt.Println("Usage: port <port>")
				continue
			}
			adress := NodeAddress("127.0.0.1:" + args[1])
			createNode(&CreateNodeArgs{adress})
			var reply string
			//time.Sleep(time.Second * 1)
			//call(adress, "Node.Ping", &HostArgs{}, &reply)
			fmt.Println("REPLY: ", reply)

		case "create":
			fmt.Println("Creates new ring")

		case "join":
			if len(args) < 2 {
				fmt.Println("Usage: join <address>")
				continue
			}
			fmt.Println("Joins existing node", args[1])

		case "quit":
			quit()

		case "host":
			fmt.Println("Hosts new node")

		case "put":
			if len(args) < 3 {
				fmt.Println("Usage: put <key> <value>")
			}
			fmt.Println("Puts key-value pair", args[1], args[2])
			putRPC(NodeAddress(args[3]), Key(args[1]), args[2])

		case "putrandom":
			fmt.Println("Puts random key-value pair")

		case "get":
			if len(args) < 2 {
				fmt.Println("Usage: get <key>")
			}
			fmt.Println("Gets value for key", args[1])
			getRPC(NodeAddress(args[2]), Key(args[1]))

		case "delete":
			if len(args) < 2 {
				fmt.Println("Usage: delete <key>")
			}
			fmt.Println("Deletes key-value pair", args[1])
			deleteRPC(NodeAddress(args[2]), Key(args[1]))

		case "clear":
			clearTerminal()

		case "dump":
			dumpRPC(NodeAddress(args[1]))

		default:
			fmt.Println("Unknown command:", commandArg)
			fmt.Println("Use 'help' for more information")
		}
	}
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("1. help - prints this help message")
	fmt.Println("2. port <n> - set the port that this node should listen on. (Default :3410)")
	fmt.Println("3. create - creates a new ring")
	fmt.Println("4. join <node> - joins an existing node")
	fmt.Println("5. quit - shut down. Ends the program.")
	fmt.Println("6. host - hosts a new node")
	fmt.Println("7. put <key> <value> - puts a key-value pair into the current ring")
	fmt.Println("8. putrandom - puts a random key-value pair")
	fmt.Println("9. get <key> - gets the value for a key")
	fmt.Println("10. delete <key> - deletes a key-value pair")
	fmt.Println("11. clear - clears the terminal screen")
	fmt.Println("12. dump - display information about the current node, used for debug")
}

func clearTerminal() {
	cmd := exec.Command("clear") // Use "clear" for Unix-like systems, "cls" for Windows
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func createRing(args *CreateArgs) *CreateReply {
	//TODO
	fmt.Println("Creates new ring")
	return &CreateReply{}
}

func joinRing(args *JoinArgs) *JoinReply {
	//TODO
	fmt.Println("Joins existing node", args.Address)
	return &JoinReply{}
}

func createNode(args *CreateNodeArgs) {
	fmt.Println("\nCreates new node", args.Address)
	CreateNode(args)
	return
}

func getNode(args *HostArgs) *HostReply {
	//TODO
	fmt.Println("Hosts new node")
	net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.ParseIP(string(args.Address))})
	return &HostReply{}
}

func quit() {
	fmt.Println("\nQuitting program.")
	os.Exit(0)
}
