package main

import "os"

func main() {
	os.Exit(1) // want "os.Exit is not allowed in main function main package"
}
