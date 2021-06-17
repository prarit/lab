package cmd

import (
	"github.com/spf13/cobra"
	lab "github.com/zaquestion/lab/internal/gitlab"
)

var projectUsers = &cobra.Command{
	Use:              "users <id>",
	Aliases:          []string{},
	Short:            "List userful information about users",
	PersistentPreRun: labPersistentPreRun,
	Run: func(cmd *cobra.Command, args []string) {
		rn, _, err := parseArgsRemoteAndID(args)
		if err != nil {
			log.Fatal(err)
		}

		p, err := lab.FindProject(rn)
		if err != nil {
			log.Fatal(err)
		}

		lab.GetProjectMembers(p.ID)
	},
}

func init() {
	projectCmd.AddCommand(projectUsers)
}
