package utils

import (
	"fmt"
	"io/fs"
	"os"
)

func Bad(format string, a ...any) string {
	msg := fmt.Sprintf(format, a...)
	return fmt.Sprint(" ", msg)
}

func Good(format string, a ...any) string {
	msg := fmt.Sprintf(format, a...)
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
func AskUserForAPath(prompt string) (string, error) {
	fmt.Print(prompt)
	var userInput string
	if _, err := fmt.Scanln(&userInput); err != nil {
		return "", err
	}
	if !fs.ValidPath(userInput) {
		return "", fmt.Errorf("Input '%v' isn't a valid path.", userInput)
	}
	return userInput, nil
}
