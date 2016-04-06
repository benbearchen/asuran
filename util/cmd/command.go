package cmd

import (
	"strings"
)

type CommandReader interface {
	Read() string
}

type Command struct {
	reader CommandReader
}

func (c *Command) OpenConsole() {
	var con ConsoleCommand
	con.Open()
	c.reader = con
}

func (c *Command) Read() string {
	return c.reader.Read()
}

func CheckCommand(input, cmd string) (string, bool) {
	arg, rest := TakeFirstArg(input)
	if arg == cmd {
		return rest, true
	}

	return "", false
}

func TakeFirstArg(cmd string) (string, string) {
	cmd = strings.TrimSpace(cmd)
	if len(cmd) == 0 {
		return "", ""
	}

	if cmd[0] != '"' {
		s := strings.IndexAny(cmd, " \t\r\n")
		if s == -1 {
			return cmd, ""
		} else {
			return cmd[:s], strings.TrimSpace(cmd[s:])
		}
	}

	for q := 1; q < len(cmd); q++ {
		if cmd[q] == '"' {
			return cmd[1:q], strings.TrimSpace(cmd[q+1:])
		}
	}

	return cmd[1:], ""
}

func SplitCommand(cmd string) []string {
	args := make([]string, 0)
	// TODO: 优化
	for len(cmd) > 0 {
		a, b := TakeFirstArg(cmd)
		if len(a) > 0 {
			args = append(args, a)
			cmd = b
		} else {
			break
		}
	}

	return args
}
