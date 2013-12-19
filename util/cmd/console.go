package cmd

import (
	"bufio"
	"os"
)

type ConsoleCommand struct {
	reader bufio.Reader
}

func (c *ConsoleCommand) Open() {
	p := bufio.NewReader(os.Stdin)
	c.reader = *p
}

func (c ConsoleCommand) Read() string {
	data, _, _ := c.reader.ReadLine()
	command := string(data)
	return command
}
