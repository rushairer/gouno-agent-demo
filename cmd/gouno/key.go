package gouno

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var keyHashCmd = &cobra.Command{
	Use:   "key-hash",
	Short: "Read a gateway API key from stdin and print a bcrypt hash",
	RunE: func(cmd *cobra.Command, args []string) error {
		value, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && len(value) == 0 {
			return fmt.Errorf("read key from stdin: %w", err)
		}
		value = strings.TrimSpace(value)
		if len(value) < 16 {
			return fmt.Errorf("gateway API key must be at least 16 characters")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(hash))
		return nil
	},
}
