#!/usr/bin/env shebang "go build -o !- !+ && !-"
// This is sort of a silly example, since "go run" is a thing.
// gccgo would probably be a better example, but whatever.

package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}
