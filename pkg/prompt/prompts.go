package prompt

import (
	"github.com/manifoldco/promptui"
)

func AskToSelect(message string, items []string) (string, error) {
	prompt := promptui.Select{
		Label: message,
		Items: items,
	}
	_, result, err := prompt.Run()

	return result, err
}

func AskForInput(message string) (string, error) {
	prompt := promptui.Prompt{
		Label: message,
	}
	return prompt.Run()
}
