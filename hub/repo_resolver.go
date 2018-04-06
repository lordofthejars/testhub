package hub

import (
	"strings"
)

var (
	github    = "github"
	gitlab    = "gitlab"
	bitbucket = "bitbucket"
	gogs      = "gogs"
	generic   = "generic"
)

func convertToCorrectBranchUrl(repoUrl, branch, repoType string) string {

	if !strings.HasSuffix(repoUrl, "/") {
		repoUrl = repoUrl + "/"
	}

	if len(repoType) > 0 {
		switch repoType {
		case github:
			return createGitHubBranchUrl(repoUrl, branch)
		case gitlab:
			return createGitlabBranchUrl(repoUrl, branch)
		case bitbucket:
			return createBitbucketBranchUrl(repoUrl, branch)
		case gogs:
			return createGogsBranchUrl(repoUrl, branch)
		}
	} else {
		switch {
		case strings.Contains(repoUrl, github):
			return createGitHubBranchUrl(repoUrl, branch)
		case strings.Contains(repoUrl, gitlab):
			return createGitlabBranchUrl(repoUrl, branch)
		case strings.Contains(repoUrl, bitbucket):
			return createBitbucketBranchUrl(repoUrl, branch)
		case strings.Contains(repoUrl, gogs):
			return createGogsBranchUrl(repoUrl, branch)
		}

	}
	return createDefaultBranchUrl(repoUrl, branch)
}

func convertToCorrectCommitUrl(repoUrl, commit, repoType string) string {

	if !strings.HasSuffix(repoUrl, "/") {
		repoUrl = repoUrl + "/"
	}

	if len(repoType) > 0 {
		switch repoType {
		case github:
			return createGitHubCommitUrl(repoUrl, commit)
		case gitlab:
			return createGitlabCommitUrl(repoUrl, commit)
		case bitbucket:
			return createBitbucketCommitUrl(repoUrl, commit)
		case gogs:
			return createGogsCommitUrl(repoUrl, commit)
		}
	} else {
		switch {
		case strings.Contains(repoUrl, github):
			return createGitHubCommitUrl(repoUrl, commit)
		case strings.Contains(repoUrl, gitlab):
			return createGitlabCommitUrl(repoUrl, commit)
		case strings.Contains(repoUrl, bitbucket):
			return createBitbucketCommitUrl(repoUrl, commit)
		case strings.Contains(repoUrl, gogs):
			return createGogsCommitUrl(repoUrl, commit)
		}

	}
	return createDefaultCommitUrl(repoUrl, commit)
}

func createDefaultBranchUrl(repoUrl, branch string) string {
	return createGitHubBranchUrl(repoUrl, branch)
}

func createGogsBranchUrl(repoUrl, branch string) string {
	return repoUrl + "src/" + branch
}

func createBitbucketBranchUrl(repoUrl, branch string) string {
	return repoUrl + "commits/branch/" + branch
}

func createGitHubBranchUrl(repoUrl, branch string) string {
	return repoUrl + "tree/" + branch
}

func createGitlabBranchUrl(repoUrl, branch string) string {
	return repoUrl + "tree/" + branch
}

func createDefaultCommitUrl(repoUrl, commit string) string {
	return createGitHubCommitUrl(repoUrl, commit)
}

func createGogsCommitUrl(repoUrl, commit string) string {
	return repoUrl + "commit/" + commit
}

func createBitbucketCommitUrl(repoUrl, commit string) string {
	return repoUrl + "commits/" + commit
}

func createGitHubCommitUrl(repoUrl, commit string) string {
	return repoUrl + "commit/" + commit
}

func createGitlabCommitUrl(repoUrl, commit string) string {
	return repoUrl + "commit/" + commit
}
