// Usage: echo 'your-password' | go run ./cmd/hashpass
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		fmt.Fprintf(os.Stderr, "read stdin: %v\n", err)
		os.Exit(1)
	}
	pw := strings.TrimSpace(line)
	if pw == "" {
		fmt.Fprintln(os.Stderr, "empty password (pipe one line on stdin)")
		os.Exit(1)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bcrypt: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(hash))
}
