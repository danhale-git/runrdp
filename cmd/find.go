package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:   "find",
	Short: "TBD",
	Long:  `TBD`,
	Args: func(cmd *cobra.Command, args []string) error {
		return cobra.RangeArgs(1, 1)(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		sortedHostKeys := configuration.HostsSortedByPattern(args[0])

		c := minInt(viper.GetInt("count"), len(sortedHostKeys))
		for i := 0; i < c; i++ {
			fmt.Printf("%d. %s\n", i+1, sortedHostKeys[i])
		}

		fmt.Print("\nEnter number to connect: ")
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		selected, err := strconv.Atoi(strings.Trim(text, "\r\n"))

		if err != nil {
			fmt.Printf("Entered value was not a whole number: %s\n", err)
			return
		}

		if selected >= c {
			fmt.Printf("Host number %d is not listed. Use -c <value> to list more hosts.\n", selected)
			return
		}

		connectToHost(sortedHostKeys[selected-1])
	},
}

func minInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func init() {
	rootCmd.AddCommand(findCmd)

	findCmd.Flags().IntP("count", "c", 6, "The number of results to display.")

	err := viper.BindPFlags(findCmd.Flags())
	if err != nil {
		panic(err)
	}
}
