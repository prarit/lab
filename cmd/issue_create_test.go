package cmd

import (
	"net/url"
	"os/exec"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_issueCreate(t *testing.T) {
	repo := copyTestRepo(t)
	cmd := exec.Command(labBinaryPath, "issue", "create", "lab-testing",
		"-m", "issue title")
	cmd.Dir = repo

	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Error creating issue: %s (%s)", string(b), err)
	}
	out := getAppOutput(b)[0]

	require.Contains(t, out, "https://gitlab.com/lab-testing/test/-/issues/")

	// Get the issue ID from the returned URL and close the issue.
	u, err := url.Parse(out)
	require.NoError(t, err, "Error parsing URL")
	id := path.Base(u.Path)

	cmd = exec.Command(labBinaryPath, "issue", "close", "lab-testing", id)
	cmd.Dir = repo
	b, err = cmd.CombinedOutput()
	require.NoError(t, err, "Error closing issue %s: %s", id, string(b))
}

func Test_issueMsg(t *testing.T) {
	tests := []struct {
		Name          string
		Msgs          []string
		ExpectedTitle string
		ExpectedBody  string
	}{
		{
			Name:          "Using messages",
			Msgs:          []string{"issue title", "issue body", "issue body 2"},
			ExpectedTitle: "issue title",
			ExpectedBody:  "issue body\n\nissue body 2",
		},
		{
			Name:          "From Editor",
			Msgs:          nil,
			ExpectedTitle: "This is the default issue template for lab",
			ExpectedBody:  "",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			test := test
			title, body, err := issueMsg("default", test.Msgs)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, test.ExpectedTitle, title)
			assert.Equal(t, test.ExpectedBody, body)
		})
	}
}

func Test_issueText(t *testing.T) {
	text, err := issueText("default")
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, `

This is the default issue template for lab
# Write a message for this issue. The first block
# of text is the title and the rest is the description.`, text)

}
