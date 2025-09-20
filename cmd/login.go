// This package contains all the cmd functions used in cli
package clicmd

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func getTokenFilePath() string {
	usr, err := user.Current()
	if err != nil {
		return "" // handle appropriately
	}
	return filepath.Join(usr.HomeDir, ".terracost_token")
}

func saveToken(token string) error {
	path := getTokenFilePath()
	return os.WriteFile(path, []byte(token), 0600)
}

func LoadToken() (string, error) {
	path := getTokenFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// Add this command to your rootCmd:
var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to TerraCost and set your auth token",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Enter your TerraCost API token: ")
		reader := bufio.NewReader(os.Stdin)
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)
		err := saveToken(token)
		if err != nil {
			fmt.Println("Failed to save token:", err)
			os.Exit(1)
		}
		fmt.Println("Login successful. Token saved.")
	},
}
