package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm asks the user a yes/no question. Returns true on y/yes.
// If autoYes is true, the question is not shown and the function returns
// true (used for scripted runs with --yes).
func Confirm(message string, autoYes bool) bool {
	if autoYes {
		fmt.Printf("%s %s\n", message, Prompt.Render("[auto-yes]"))
		return true
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s %s ", message, Prompt.Render("[y/n]"))

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	switch response {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		fmt.Printf("Please answer %s or %s.", Prompt.Render("[y/yes]"), Prompt.Render("[n/no]"))
		return Confirm(message, autoYes)
	}
}
