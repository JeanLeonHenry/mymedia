package utils

import (
	"fmt"
	"os"
)

var Abs = func(x int) int {
	if x <= 0 {
		return -x
	}
	return x
}

func AcceptOrQuit(prompt string) {
	fmt.Print(prompt + " [y/N] ")
	var userInput string
	if _, err := fmt.Scanln(&userInput); err != nil || userInput != "y" {
		fmt.Println("Quitting.")
		os.Exit(1)
	}
}
