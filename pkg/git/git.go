package git

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/msoovali/jira-release-maker/pkg/entity"
)

const (
	PROJECTS_DIR          = "projects"
	RELEASE_BRANCH_PREFIX = "origin/release/"
)

type git struct {
	path       string
	project    entity.Project
	projectDir string
}

func LoadRepository(project entity.Project) (*git, error) {
	gitExecutable, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git executable not found from path")
	}
	git := &git{
		path:       gitExecutable,
		project:    project,
		projectDir: PROJECTS_DIR + "/" + project.Name + "/",
	}
	if !git.repositoryExists() {
		if err = git.cloneRepository(); err != nil {
			return nil, err
		}
	}

	return git, nil
}

func (git git) cloneRepository() error {
	cmdGitClone := git.buildGitCmd([]string{"clone", git.project.GitUrl, git.projectDir}, false)
	return runCommand(*cmdGitClone)
}

func (git git) repositoryExists() bool {
	_, err := os.Stat(git.projectDir + ".git/")
	if os.IsNotExist(err) {
		_, err := os.Stat(git.projectDir)
		if os.IsNotExist(err) {
			return false
		}
		fmt.Printf("Project dir (%s) exists but is not a git repository. Removing!\n", git.projectDir)
		os.RemoveAll(git.projectDir)
		return false
	}

	return true
}

func (git git) GetLatestTag() string {
	git.fetch()
	git.checkout(git.project.MainBranch)
	git.pull()
	cmdGitGetLatestTag := git.buildGitCmd([]string{"describe", "--tags", "--abbrev=0"}, true)
	logCommand(*cmdGitGetLatestTag)
	byteValue, err := cmdGitGetLatestTag.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	return string(byteValue)
}

func (git git) buildGitCmd(args []string, executeInProjectDir bool) *exec.Cmd {
	gitBaseCmd := []string{git.path}
	if executeInProjectDir {
		gitBaseCmd = append(gitBaseCmd, []string{"-C", git.projectDir}...)
	}
	cmdGit := &exec.Cmd{
		Path:   git.path,
		Args:   append(gitBaseCmd, args...),
		Stderr: os.Stdout,
	}

	return cmdGit
}

func runCommand(command exec.Cmd) error {
	logCommand(command)
	if err := command.Run(); err != nil {
		fmt.Println("Error: ", err)
		return err
	}

	return nil
}

func logCommand(command exec.Cmd) {
	fmt.Printf("Running command %s\n", command.Args)
}

func (git git) checkout(branch string) {
	cmdGitCheckout := git.buildGitCmd([]string{"checkout", "-f", branch}, true)
	runCommand(*cmdGitCheckout)
}

func (git git) pull() {
	cmdGitPull := git.buildGitCmd([]string{"pull"}, true)
	runCommand(*cmdGitPull)
}

func (git git) fetch() {
	cmdGitFetch := git.buildGitCmd([]string{"fetch"}, true)
	runCommand(*cmdGitFetch)
}

func (git git) FindPossibleReleaseBranches(releaseName string) []string {
	var possibleReleaseBranches []string
	now := time.Now()
	numberedReleaseBranch := RELEASE_BRANCH_PREFIX + releaseName
	dateReleaseBranchFullYear := RELEASE_BRANCH_PREFIX + fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day())
	dateReleaseBranch2DigitYear := RELEASE_BRANCH_PREFIX + fmt.Sprintf("%s-%02d-%02d", strconv.Itoa(now.Year())[2:4], now.Month(), now.Day())
	cmdGitBranchFind := git.buildGitCmd([]string{"branch", "-a", "-l", numberedReleaseBranch, dateReleaseBranchFullYear, dateReleaseBranch2DigitYear}, true)
	logCommand(*cmdGitBranchFind)
	byteValue, err := cmdGitBranchFind.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return possibleReleaseBranches
	}
	trimmedReleaseBranchesString := strings.TrimSpace(string(byteValue))
	if trimmedReleaseBranchesString == "" {
		return possibleReleaseBranches
	}
	releaseBranches := strings.Split(trimmedReleaseBranchesString, " ")
	for _, releaseBranch := range releaseBranches {
		possibleReleaseBranches = append(possibleReleaseBranches, strings.TrimSpace(releaseBranch))
	}

	return possibleReleaseBranches
}

func (git git) FindGitLogDiffWithMaster(releaseBranch string) string {
	releaseBranch = strings.TrimSpace(strings.ReplaceAll(releaseBranch, "remotes/origin/", ""))
	git.checkout(releaseBranch)
	git.pull()
	cmdGitLog := git.buildGitCmd([]string{"log", "--pretty=format:%s", "--no-merges", "--reverse", releaseBranch + ".." + git.project.MainBranch}, true)
	logCommand(*cmdGitLog)
	byteValue, err := cmdGitLog.Output()
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	return string(byteValue)
}

func (git git) FindIssuesInRelease(gitLogOutput string) []string {
	issueKeyAdded := make(map[string]bool)
	var issues []string
	commitMessages := strings.Split(gitLogOutput, "\n")
	re := regexp.MustCompile(`^([A-Z]+-[0-9]+(\s+|,)*)+`)
	for _, message := range commitMessages {
		result := strings.TrimSpace(string(re.Find([]byte(message))))
		if result != "" {
			for _, key := range strings.Split(result, " ") {
				if !issueKeyAdded[key] {
					issueKeyAdded[key] = true
					issues = append(issues, key)
				}
			}
		}
	}

	return issues
}
