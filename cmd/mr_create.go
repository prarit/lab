package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
	gitlab "github.com/xanzy/go-gitlab"
	"github.com/zaquestion/lab/internal/action"
	"github.com/zaquestion/lab/internal/git"
	lab "github.com/zaquestion/lab/internal/gitlab"
)

// mrCmd represents the mr command
var mrCreateCmd = &cobra.Command{
	Use:              "create [remote [branch]]",
	Aliases:          []string{"new"},
	Short:            "Open a merge request on GitLab",
	Long:             `Creates a merge request (default: MR created on default branch of origin)`,
	Args:             cobra.MaximumNArgs(2),
	PersistentPreRun: LabPersistentPreRun,
	Run:              runMRCreate,
}

func init() {
	mrCreateCmd.Flags().StringArrayP("message", "m", []string{}, "use the given <msg>; multiple -m are concatenated as separate paragraphs")
	mrCreateCmd.Flags().StringSliceP("assignee", "a", []string{}, "set assignee by username; can be specified multiple times for multiple assignees")
	mrCreateCmd.Flags().StringSliceP("label", "l", []string{}, "add label <label>; can be specified multiple times for multiple labels")
	mrCreateCmd.Flags().BoolP("remove-source-branch", "d", false, "remove source branch from remote after merge")
	mrCreateCmd.Flags().BoolP("squash", "s", false, "squash commits when merging")
	mrCreateCmd.Flags().Bool("allow-collaboration", false, "allow commits from other members")
	mrCreateCmd.Flags().String("milestone", "", "set milestone by milestone title or ID")
	mrCreateCmd.Flags().StringP("file", "F", "", "use the given file as the Description")
	mrCreateCmd.Flags().Bool("force-linebreak", false, "append 2 spaces to the end of each line to force markdown linebreaks")
	mrCreateCmd.Flags().BoolP("cover-letter", "c", false, "do not comment changelog and diffstat")
	mrCreateCmd.Flags().Bool("draft", false, "mark the merge request as draft")
	mergeRequestCmd.Flags().AddFlagSet(mrCreateCmd.Flags())

	mrCmd.AddCommand(mrCreateCmd)

	carapace.Gen(mrCreateCmd).FlagCompletion(carapace.ActionMap{
		"label": carapace.ActionMultiParts(",", func(c carapace.Context) carapace.Action {
			if project, _, err := parseArgsRemoteAndProject(c.Args); err != nil {
				return carapace.ActionMessage(err.Error())
			} else {
				return action.Labels(project).Invoke(c).Filter(c.Parts).ToA()
			}
		}),
		"milestone": carapace.ActionCallback(func(c carapace.Context) carapace.Action {
			if project, _, err := parseArgsRemoteAndProject(c.Args); err != nil {
				return carapace.ActionMessage(err.Error())
			} else {
				return action.Milestones(project, action.MilestoneOpts{Active: true})
			}
		}),
	})

	carapace.Gen(mrCreateCmd).PositionalCompletion(
		action.Remotes(),
		action.RemoteBranches(0),
	)
}

// getAssignee returns the assigneeID for use with other GitLab API calls.
// NOTE: It is also used by issue_create.go
func getAssigneeID(assignee string) *int {
	if assignee == "" {
		return nil
	}
	if assignee[0] == '@' {
		assignee = assignee[1:]
	}
	assigneeID, err := lab.UserIDFromUsername(assignee)
	if err != nil {
		return nil
	}
	if assigneeID == -1 {
		return nil
	}
	return gitlab.Int(assigneeID)
}

// getAssignees returns the assigneeIDs for use with other GitLab API calls.
func getAssigneeIDs(assignees []string) []int {
	var ids []int
	for _, a := range assignees {
		ids = append(ids, *getAssigneeID(a))
	}
	return ids
}

func runMRCreate(cmd *cobra.Command, args []string) {
	msgs, err := cmd.Flags().GetStringArray("message")
	if err != nil {
		log.Fatal(err)
	}
	assignees, err := cmd.Flags().GetStringSlice("assignee")
	if err != nil {
		log.Fatal(err)
	}
	filename, err := cmd.Flags().GetString("file")
	if err != nil {
		log.Fatal(err)
	}
	coverLetterFormat, err := cmd.Flags().GetBool("cover-letter")
	if err != nil {
		log.Fatal(err)
	}
	branch, err := git.CurrentBranch()
	if err != nil {
		log.Fatal(err)
	}

	sourceRemote := determineSourceRemote(branch)
	sourceProjectName, err := git.PathWithNameSpace(sourceRemote)
	if err != nil {
		log.Fatal(err)
	}

	remoteBranch, err := git.CurrentUpstreamBranch()
	if remoteBranch == "" {
		// Fall back to local branch
		remoteBranch, err = git.CurrentBranch()
	}

	if err != nil {
		log.Fatal(err)
	}

	p, err := lab.FindProject(sourceProjectName)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := lab.GetCommit(p.ID, remoteBranch); err != nil {
		err = errors.Wrapf(
			err,
			"aborting MR, source branch %s not present on remote %s. did you forget to push?",
			remoteBranch, sourceRemote)
		log.Fatal(err)
	}

	targetRemote := defaultRemote
	if len(args) > 0 {
		targetRemote = args[0]
		ok, err := git.IsRemote(targetRemote)
		if err != nil || !ok {
			log.Fatal(errors.Wrapf(err, "%s is not a valid remote", targetRemote))
		}
	}
	targetProjectName, err := git.PathWithNameSpace(targetRemote)
	if err != nil {
		log.Fatal(err)
	}
	targetProject, err := lab.FindProject(targetProjectName)
	if err != nil {
		log.Fatal(err)
	}

	targetBranch := targetProject.DefaultBranch
	if len(args) > 1 && targetBranch != args[1] {
		targetBranch = args[1]
		if _, err := lab.GetCommit(targetProject.ID, targetBranch); err != nil {
			err = errors.Wrapf(
				err,
				"aborting MR, %s:%s is not a valid target. Did you forget to push %s to %s?",
				targetRemote, branch, branch, targetRemote)
			log.Fatal(err)
		}
	}

	labelTerms, err := cmd.Flags().GetStringSlice("label")
	if err != nil {
		log.Fatal(err)
	}
	labels, err := MapLabels(targetProjectName, labelTerms)
	if err != nil {
		log.Fatal(err)
	}

	milestoneArg, _ := cmd.Flags().GetString("milestone")
	milestoneID, _ := strconv.Atoi(milestoneArg)

	var milestone *int
	if milestoneID > 0 {
		milestone = &milestoneID
	} else if milestoneArg != "" {
		ms, err := lab.MilestoneGet(targetProjectName, milestoneArg)
		if err != nil {
			log.Fatal(err)
		}
		milestone = &ms.ID
	} else {
		milestone = nil
	}

	var title, body string

	if filename != "" {
		if len(msgs) > 0 || coverLetterFormat {
			log.Fatal("option -F cannot be combined with -m/-c")
		}

		title, body, err = editGetTitleDescFromFile(filename)
		if err != nil {
			log.Fatal(err)
		}
	} else if len(msgs) > 0 {
		if coverLetterFormat {
			log.Fatal("option -m cannot be combined with -c/-F")
		}

		title, body = msgs[0], strings.Join(msgs[1:], "\n\n")
	} else {
		msg, err := mrText(targetBranch, branch, sourceRemote, targetRemote, coverLetterFormat)
		if err != nil {
			log.Fatal(err)
		}

		title, body, err = git.Edit("MERGEREQ", msg)
		if err != nil {
			_, f, l, _ := runtime.Caller(0)
			log.Fatal(f+":"+strconv.Itoa(l)+" ", err)
		}
	}

	if title == "" {
		log.Fatal("aborting MR due to empty MR msg")
	}

	linebreak, _ := cmd.Flags().GetBool("force-linebreak")
	if linebreak {
		body = textToMarkdown(body)
	}

	draft, _ := cmd.Flags().GetBool("draft")
	if draft {
		isWIP := hasPrefix(title, "wip:")
		isDraft := hasPrefix(title, "draft:") ||
			hasPrefix(title, "[draft]") ||
			hasPrefix(title, "(draft)")

		if !isWIP && !isDraft {
			title = "Draft: " + title
		}
	}

	removeSourceBranch, _ := cmd.Flags().GetBool("remove-source-branch")
	squash, _ := cmd.Flags().GetBool("squash")
	allowCollaboration, _ := cmd.Flags().GetBool("allow-collaboration")

	mrURL, err := lab.MRCreate(sourceProjectName, &gitlab.CreateMergeRequestOptions{
		SourceBranch:       &branch,
		TargetBranch:       gitlab.String(targetBranch),
		TargetProjectID:    &targetProject.ID,
		Title:              &title,
		Description:        &body,
		AssigneeIDs:        getAssigneeIDs(assignees),
		RemoveSourceBranch: &removeSourceBranch,
		Squash:             &squash,
		AllowCollaboration: &allowCollaboration,
		Labels:             labels,
		MilestoneID:        milestone,
	})
	if err != nil {
		// FIXME: not exiting fatal here to allow code coverage to
		// generate during Test_mrCreate. In the meantime API failures
		// will exit 0
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Println(mrURL + "/diffs")
}

func mrText(base, head, sourceRemote, targetRemote string, coverLetterFormat bool) (string, error) {
	var (
		commitMsg string
		err       error
	)
	remoteBase := fmt.Sprintf("%s/%s", targetRemote, base)
	targetBase := fmt.Sprintf("%s/%s", sourceRemote, head)
	commitMsg = ""

	numCommits := git.NumberCommits(remoteBase, targetBase)
	if numCommits == 1 {
		commitMsg, err = git.LastCommitMessage()
		if err != nil {
			return "", err
		}
	}
	if numCommits == 0 {
		return "", fmt.Errorf("Aborting: The resulting Merge Request from %s to %s has 0 commits.", remoteBase, targetBase)
	}

	const tmpl = `{{if .InitMsg}}{{.InitMsg}}{{end}}

{{if .Tmpl}}{{.Tmpl}}{{end}}
{{.CommentChar}} Requesting a merge into {{.Base}} from {{.Head}} ({{.NumCommits}} commits)
{{.CommentChar}}
{{.CommentChar}} Write a message for this merge request. The first block
{{.CommentChar}} of text is the title and the rest is the description.{{if .CommitLogs}}
{{.CommentChar}}
{{.CommentChar}} Changes:
{{.CommentChar}}
{{.CommitLogs}}{{end}}`

	mrTmpl := lab.LoadGitLabTmpl(lab.TmplMR)

	commitLogs, err := git.Log(remoteBase, targetBase)
	if err != nil {
		return "", err
	}

	commitLogs = strings.TrimSpace(commitLogs)
	commentChar := git.CommentChar()

	if !coverLetterFormat {
		startRegexp := regexp.MustCompilePOSIX("^")
		commitLogs = startRegexp.ReplaceAllString(commitLogs, fmt.Sprintf("%s ", commentChar))
	} else {
		commitLogs = "\n" + strings.TrimSpace(commitLogs)
	}

	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return "", err
	}

	msg := &struct {
		InitMsg     string
		Tmpl        string
		CommentChar string
		Base        string
		Head        string
		CommitLogs  string
		NumCommits  int
	}{
		InitMsg:     commitMsg,
		Tmpl:        mrTmpl,
		CommentChar: commentChar,
		Base:        targetRemote + ":" + base,
		Head:        sourceRemote + ":" + head,
		CommitLogs:  commitLogs,
		NumCommits:  numCommits,
	}

	var b bytes.Buffer
	err = t.Execute(&b, msg)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
