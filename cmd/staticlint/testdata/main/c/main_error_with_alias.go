package main

import exampleos "os"

func main() {
	exampleos.Exit(1) // want "os.Exit is not allowed in main function main package"
}
