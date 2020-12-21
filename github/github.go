package github

import (
	"context"
	"os"
	"strconv"

	"github.com/google/go-github/v33/github"
	"github.com/lorislab/changelog/changelog"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type gitHubItem struct {
	Issue *github.Issue
}

// GetID of the github issue
func (g gitHubItem) GetID() string {
	return strconv.Itoa(g.Issue.GetNumber())
}

// GetURL of the github issue
func (g gitHubItem) GetURL() string {
	return g.Issue.GetURL()
}

// GetTitle of the github issue
func (g gitHubItem) GetTitle() string {
	return g.Issue.GetTitle()
}

// GithubClientService github release service
type githubClientService struct {
	client *github.Client
	ctx    context.Context
	repo   string
	owner  string
}

// CreateClient create client
func CreateClient(owner, repo, token string) changelog.ClientService {
	r := githubClientService{
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
func (g githubClientService) FindVersionIssues(version string, groups []*changelog.Group) {
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
		log.Warnf("Version %s not found", version)
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

	for _, issue := range issues {
		if issue.GetState() == "open" {
			log.Warnf("Open issue #%d %s", issue.GetNumber(), issue.GetURL())
		}

		labels := createSet(issue)
		for _, group := range groups {
			if group.ContaintsLabels(labels) {
				group.Items = append(group.Items, gitHubItem{Issue: issue})
			}
		}
	}
}

// CreateRelease create release
func (g githubClientService) CreateRelease(version string, prerelease bool, output string) {
	release := github.RepositoryRelease{
		TagName:    &version,
		Name:       &version,
		Prerelease: &prerelease,
		Body:       &output,
	}
	g.client.Repositories.CreateRelease(g.ctx, g.owner, g.repo, &release)
}

// create set of labels
func createSet(g *github.Issue) map[string]bool {
	labels := make(map[string]bool)
	for _, label := range g.Labels {
		labels[label.GetName()] = true
	}
	return labels
}
