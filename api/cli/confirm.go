package cli

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// Confirm starts a rollback db cli application
func Confirm(message string) (bool, error) {
	log.Printf("%s [y/N]", message)

	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	answer = strings.Replace(answer, "\n", "", -1)
	answer = strings.ToLower(answer)

	return answer == "y" || answer == "yes", nil

}
