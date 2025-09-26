package main

import "os"

func main() {
	os.Exit(1) // want "os.Exit usage in main function is forbidden. Use return instead of os.Exit"
}
