package main

import (
	"fmt"

	conf "github.com/msoovali/jira-release-maker/pkg/config"
	"github.com/msoovali/jira-release-maker/pkg/git"
	"github.com/msoovali/jira-release-maker/pkg/jira"
	"github.com/msoovali/jira-release-maker/pkg/prompt"
	"github.com/msoovali/jira-release-maker/pkg/versionprediction"
)

const (
	CUSTOM                 = "custom"
	JIRA_SUMMARY_SEPARATOR = "_"
)

func main() {
	// load settings
	settings, err := conf.LoadConfig()
	if err != nil {
		fmt.Println("Failed to load settings: ", err)
		return
	}
	// ask for project to release
	projectName, err := prompt.AskToSelect("Select project You would like to release", settings.GetProjectsNames())
	if err != nil {
		return
	}
	project := settings.GetProjectByName(projectName)
	// load git repository for project
	gitRepo, err := git.LoadRepository(*project)
	if err != nil {
		fmt.Printf("Repository loading failed! %v\n", err)
		return
	}
	// get latest tag for next version prediction
	latestTag := gitRepo.GetLatestTag()
	versions := append(versionprediction.GetVersionPredictions(latestTag), CUSTOM)
	nextVersion, err := prompt.AskToSelect("Select next version number", versions)
	if err != nil {
		return
	}
	if nextVersion == CUSTOM {
		nextVersion, err = prompt.AskForInput("Enter next version number (example 2.22.1)")
		if err != nil {
			return
		}
	}
	releaseName := project.Name + JIRA_SUMMARY_SEPARATOR + nextVersion
	jiraClient := jira.NewClient(settings.ApiToken, settings.AtlassianDomain)
	// check if release with given name already exists
	existingReleaseKey, err := jiraClient.ReleaseExists(releaseName)
	if err != nil {
		fmt.Println("Failed to determine if release exists in Jira: ", err)
		return
	}
	if existingReleaseKey != "" {
		fmt.Printf("Release %s already exists in jira: https://%s.atlassian.net/browse/%s\n", releaseName, settings.AtlassianDomain, existingReleaseKey)
		return
	}
	// determine release branch
	possibleReleaseBranches := append(gitRepo.FindPossibleReleaseBranches(releaseName), CUSTOM)
	releaseBranch, err := prompt.AskToSelect("Select release branch", possibleReleaseBranches)
	if err != nil {
		return
	}
	if releaseBranch == CUSTOM {
		releaseBranch, err = prompt.AskForInput(fmt.Sprintf("Enter release branch (example %s)", project.ReleaseBranchExample))
		if err != nil {
			return
		}
	}
	// calculate diff and issue keys
	logDiff := gitRepo.FindGitLogDiffWithMaster(releaseBranch)
	issueKeys := gitRepo.FindIssuesInRelease(logDiff)
	if len(issueKeys) < 1 {
		fmt.Printf("Issue list is empty! Not making release task!\n")
		return
	}
	fmt.Printf("Commit list:\n%s\n", logDiff)
	fmt.Printf("Issue list:\n%v\n", issueKeys)
	releaseKey, err := jiraClient.CreateInitialTestRelease(*project, releaseName)
	if err != nil {
		fmt.Println("Issue creation failed. ", err)
		return
	}

	fmt.Printf("Release created succesfully: https://%s.atlassian.net/browse/%s\n", settings.AtlassianDomain, releaseKey)
}
