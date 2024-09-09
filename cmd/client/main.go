package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/KoduIsGreat/knight-game/networking/client"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error running client: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	c, err := client.Dial(":1111")
	if err != nil {
		return err
	}
	defer c.Close()

	fmt.Println("Connected to server")
	msg := "Hello, server!"
	c.Write([]byte(msg))

	for {
		var msg string
		fmt.Print("Enter message: ")
		if _, err := fmt.Fscanf(os.Stdin, "%s", &msg); err != nil {
			return err
		}

		if _, err := client.Write(c, []byte(msg)); err != nil {
			return err
		}

		data, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, ">[%d len] %s\n", len(data), data)

	}
}
