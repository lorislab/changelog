package github

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/google/go-github/v33/github"
	"github.com/lorislab/changelog/changelog"
	"golang.org/x/oauth2"
)

type GitHubItem struct {
	Issue *github.Issue
}

// GetID of the github issue
func (g GitHubItem) GetID() string {
	return strconv.Itoa(g.Issue.GetNumber())
}

// GetURL of the github issue
func (g GitHubItem) GetURL() string {
	return g.Issue.GetURL()
}

// GetTitle of the github issue
func (g GitHubItem) GetTitle() string {
	return g.Issue.GetTitle()
}

// GithubClientService github release service
type GithubClientService struct {
	client *github.Client
	ctx    context.Context
	repo   string
	owner  string
}

// CreateClient create client
func CreateClient(owner, repo, token string) *GithubClientService {
	r := GithubClientService{
		repo:  repo,
		owner: owner,
		ctx:   context.Background(),
	}

	if len(token) > 0 {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(r.ctx, ts)
		r.client = github.NewClient(tc)
	}
	return &r
}

// FindVersionIssues find issues for the release
func (g GithubClientService) FindVersionIssues(version string, groups []*changelog.Group) {
	options := github.MilestoneListOptions{
		State: "open",
	}
	milestones, _, err := g.client.Issues.ListMilestones(g.ctx, g.owner, g.repo, &options)
	if err != nil {
		panic(err)
	}
	var milestone *github.Milestone
	for _, m := range milestones {
		if m.GetTitle() == version {
			milestone = m
			break
		}
	}
	if milestone == nil {
		fmt.Printf("Version %s not found\n", version)
		os.Exit(1)
	}

	issueList := github.IssueListByRepoOptions{
		Milestone: strconv.Itoa(milestone.GetNumber()),
		State:     "all",
	}
	// check reponse for all issues
	issues, _, err := g.client.Issues.ListByRepo(g.ctx, g.owner, g.repo, &issueList)
	if err != nil {
		panic(err)
	}

	type void struct{}
	var member void

	for _, issue := range issues {
		if issue.GetState() == "open" {
			fmt.Printf("Warning find open issue #%d %s\n", issue.GetNumber(), issue.GetURL())
		}
		set := make(map[string]void)
		for _, label := range issue.Labels {
			set[label.GetName()] = member
		}

		for _, group := range groups {
			for _, label := range group.Labels {
				_, exists := set[label]
				if exists {
					group.Items = append(group.Items, GitHubItem{Issue: issue})
				}
			}
		}
	}
}

// CreateRelease create release
func (g GithubClientService) CreateRelease(version string, prerelease bool, output string) {
	release := github.RepositoryRelease{
		TagName:    &version,
		Name:       &version,
		Prerelease: &prerelease,
		Body:       &output,
	}
	g.client.Repositories.CreateRelease(g.ctx, g.owner, g.repo, &release)
}
