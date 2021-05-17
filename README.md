# Jira release task maker - automation script
Release maker helps You create Jira issues for projects. Just define Your Jira release task layout in `jira-create-issue.json` file, definitions can be found [here](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issues/#api-rest-api-3-issue-post)
## Config and requirements
You need to have `git` available from path.
Make sure in conf folder You have `settings.json` file with following structure. Define as many projects You need.
```json
{
  "atlassian_domain": "your-atlassian-domain",
  "projects": [
    {
      "name": "how-you-call-project1",
      "jira_component_name": "project1-jira-component",
      "git_url": "git@bitbucket.org:project1-repository.git",
      "release_branch_example": "release-branch-naming-example",
      "main_branch": "master"
    },
    {
      "name": "how-you-call-project2",
      "jira_component_name": "project2-jira-component",
      "git_url": "git@bitbucket.org:project2-repository.git",
      "release_branch_example": "release-branch-naming-example",
      "main_branch": "master"
    }
  ]
}
```
## How to run?
In case You have go installed `go run cmd/cli/*.go`
If You don't have go installed then there are prebuilt executables available in `builds` directory.
Copy Your environment specific executable into project root dir, so that `conf` directory is accessible for application, then execute with `./your-executable` or on windows with `start your-executable.exe`