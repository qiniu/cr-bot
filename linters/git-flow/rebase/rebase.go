package rebase

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cr-bot/config"
	"github.com/cr-bot/linters"
	"github.com/google/go-github/v57/github"
	"github.com/qiniu/x/log"
)

var lintName = "rebase-suggestion"

func init() {
	// register linter
	linters.RegisterCommentHandler(lintName, rebaseSuggestionHandler)
}

func rebaseSuggestionHandler(linterConfig config.Linter, agent linters.Agent, event github.PullRequestEvent) error {
	opts := &github.ListOptions{}
	commits, response, err := agent.GitHubClient().PullRequests.ListCommits(context.Background(), event.GetRepo().GetOwner().GetLogin(), event.GetRepo().GetName(), event.GetNumber(), opts)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		log.Errorf("list commits failed: %v", response)
		return fmt.Errorf("list commits failed: %v", response.Body)
	}

	comment := checkCommitMessage(commits)
	if len(comment) > 0 {
		return agent.CreateComment(event, comment)
	}

	return nil
}

func checkCommitMessage(commits []*github.RepositoryCommit) string {
	pattern := `^Merge (.*) into (.*)$`
	reg := regexp.MustCompile(pattern)

	for _, commit := range commits {
		if commit.Commit != nil && commit.Commit.Message != nil {
			if reg.MatchString(*commit.Commit.Message) {
				return "please rebase your PR"
			}
		}
	}

	return ""
}
