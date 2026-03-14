package util

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

func Kubectl(args string) (string, error) {
	argv, err := SplitArgs(args)
	if err != nil {
		return "", fmt.Errorf("failed to parse argument vector: %v", err)
	}
	cmd := exec.Command("kubectl", argv...)
	cmd.Env = os.Environ()
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}
	res, err := cmd.CombinedOutput()
	return string(res), err
}

func SplitArgs(cmd string) ([]string, error) {
	var args []string
	var current strings.Builder

	inSingle := false
	inDouble := false
	escaped := false

	flush := func() {
		if current.Len() > 0 {
			args = append(args, current.String())
			current.Reset()
		}
	}

	for _, r := range cmd {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false

		case r == '\\':
			escaped = true

		case r == '\'' && !inDouble:
			inSingle = !inSingle

		case r == '"' && !inSingle:
			inDouble = !inDouble

		case r == '|' && !inSingle && !inDouble:
			flush()
			args = append(args, "|")

		case unicode.IsSpace(r) && !inSingle && !inDouble:
			flush()

		default:
			current.WriteRune(r)
		}
	}

	if escaped {
		return nil, fmt.Errorf("unfinished escape sequence")
	}
	if inSingle || inDouble {
		return nil, fmt.Errorf("unclosed quote in command")
	}

	flush()
	return args, nil
}
