package main

import myos "os"

func main() {
	myos.Exit(0) // want "os.Exit in main function of main package is prohibited"
}
