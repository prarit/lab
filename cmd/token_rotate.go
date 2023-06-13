package cmd

import (
	"strconv"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"

	lab "github.com/zaquestion/lab/internal/gitlab"
)

var tokenRotateCmd = &cobra.Command{
	Use:   "rotate [token ID]",
	Short: "rotate a Personal Access Token",
	Args:  cobra.MaximumNArgs(1),
	Example: heredoc.Doc(`
		lab token rotate`),

	PersistentPreRun: labPersistentPreRun,
	Run: func(cmd *cobra.Command, args []string) {
		id := 0
		if len(args) == 1 {
			var err error
			id, err = strconv.Atoi(args[0])
			if err != nil {
				log.Fatal(err)
			}
		} else {
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				log.Fatalf("Must specify a valid token ID or name\n")
			}
			PATs, err := lab.GetAllPATs()
			if err != nil {
				log.Fatal(err)
			}
			for _, PAT := range PATs {
				if PAT.Name == name {
					id = PAT.ID
					break
				}
			}
			log.Fatalf("%s is not a valid Token name\n", name)
		}

		if id == 0 {
			log.Fatalf("Must specify a valid token ID or name\n")
		}

		PAT, err := lab.RotatePAT(id)
		if err != nil {
			log.Fatal(err)
		}

		dumpToken(PAT)
	},
}

func init() {
	tokenRotateCmd.Flags().StringP("name", "", "", "name of token (can be obtained from 'lab token list'")
	tokenCmd.AddCommand(tokenRotateCmd)
}
