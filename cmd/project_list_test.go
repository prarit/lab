package cmd

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_projectList(t *testing.T) {
	repo := copyTestRepo(t)
	cmd := exec.Command(labBinaryPath, "project", "list", "-m")
	cmd.Dir = repo

	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	projects := strings.Split(string(b), "\n")
	t.Log(projects)
	require.Equal(t, "lab-testing/www-gitlab-com", projects[0])
}

func Test_projectList_many(t *testing.T) {
	repo := copyTestRepo(t)
	cmd := exec.Command(labBinaryPath, "project", "list", "-n", "101")
	cmd.Dir = repo

	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	projects := getAppOutput(b)
	assert.Equal(t, 101, len(projects), "Expected 101 projects listed")
	assert.NotContains(t, projects, "PASS")
}
