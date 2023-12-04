package main

import (
	"bufio"
	"fmt"
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
			fmt.Print("Port: " + args[1])
		case "host":
			fmt.Println("Hosts new node")
		case "join":
			if len(args) < 2 {
				fmt.Println("Usage: join <node>")
				continue
			}
			fmt.Println("Joins existing node", args[1])
		case "quit":
			fmt.Println("Quitting program.")
			return
		case "put":
			if len(args) < 3 {
				fmt.Println("Usage: put <key> <value>")
				continue
			}
			fmt.Println("Puts key-value pair", args[1], args[2])
		case "putrandom":
			fmt.Println("Puts random key-value pair")
		case "get":
			if len(args) < 2 {
				fmt.Println("Usage: get <key>")
				continue
			}
			fmt.Println("Gets value for key", args[1])
		case "delete":
			if len(args) < 2 {
				fmt.Println("Usage: delete <key>")
				continue
			}
			fmt.Println("Deletes key-value pair", args[1])
		case "clear":
			clearTerminal()
		default:
			fmt.Println("Unknown command:", commandArg)
			fmt.Println("Use 'help' for more information")
		}
	}
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("1. help - prints this help message")
	fmt.Println("2. port <port> - prints the port number of the node")
	fmt.Println("3. host - hosts a new node")
	fmt.Println("4. join <node> - joins an existing node")
	fmt.Println("5. quit - quits all nodes")
	fmt.Println("6. put <key> <value> - puts a key-value pair")
	fmt.Println("7. putrandom - puts a random key-value pair")
	fmt.Println("8. get <key> - gets the value for a key")
	fmt.Println("9. delete <key> - deletes a key-value pair")
	fmt.Println("10. clear - clears the terminal screen")
}

func clearTerminal() {
	cmd := exec.Command("clear") // Use "clear" for Unix-like systems, "cls" for Windows
	cmd.Stdout = os.Stdout
	cmd.Run()
}
