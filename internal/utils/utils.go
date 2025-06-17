package utils

import (
	"fmt"
	"os"
)

func Bad(msg string) string {
	return fmt.Sprint(" ", msg)
}

func Good(msg string) string {
	return fmt.Sprint("✓ ", msg)
}

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
func AskUser(prompt string) bool {
	fmt.Print(prompt + " [y/N] ")
	var userInput string
	if _, err := fmt.Scanln(&userInput); err != nil || userInput != "y" {
		return false
	}
	return true
}
