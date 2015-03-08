package term

import (
	"fmt"
)

func Password(prompt string) (string, error) {
	fmt.Print(prompt)
	var pass string
	var c rune
	for c != '\n' {
		_, err := fmt.Scanf("%c", &c)
		if err != nil {
			return "", err
		}
		if c != '\n' {
			pass = fmt.Sprintf("%s%c", pass, c)
		}
		fmt.Print("\b")
	}
	return pass, nil
}
