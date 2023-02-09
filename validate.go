package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	diffData, err := ioutil.ReadFile("projects.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	fmt.Print(diffData)

	os.Exit(0)
}
