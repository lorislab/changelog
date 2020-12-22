package github

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v33/github"
	changelog "github.com/lorislab/changelog/api"
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

// Init initialize github client and configuration
func Init(repository, token string) (changelog.ClientService, string) {
	if len(token) == 0 {
		log.Fatal("Github token is mandatory")
	}

	ver := ""
	repo := repository

	// check github actions environment variables
	log.WithFields(log.Fields{
		"actions":    os.Getenv("GITHUB_ACTIONS"),
		"version":    os.Getenv("GITHUB_REF"),
		"repository": os.Getenv("GITHUB_REPOSITORY"),
	}).Info("Github actions")

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		ver = os.Getenv("GITHUB_REF")
		log.WithField("version", ver).Debug("Setup the version from GITHUB_REF env variable")
		if len(repo) == 0 {
			repo = os.Getenv("GITHUB_REPOSITORY")
			log.WithField("repository", repo).Debug("Setup the version from GITHUB_REPOSITORY env variable")
		}
	}

	// check repository
	if len(repo) == 0 {
		log.WithFields(log.Fields{
			"input":   repository,
			"current": repo,
		}).Fatal("Repository is empty")
	}

	// create github client
	client := createClient(repository, token)
	return client, ver
}

func createClient(repository, token string) changelog.ClientService {
	items := strings.Split(repository, "/")
	if len(items) != 2 {
		log.WithField("repository", repository).Fatal("Wrong format for the github repository owner/repo")
	}
	result := githubClientService{
		owner: items[0],
		repo:  items[1],
		ctx:   context.Background(),
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(result.ctx, ts)
	result.client = github.NewClient(tc)

	return &result
}

// FindVersionIssues find issues for the release
func (g githubClientService) FindVersionIssues(version string, sections []*changelog.Section) {

	milestone := g.findMilstone(version)
	issueList := github.IssueListByRepoOptions{
		Milestone: strconv.Itoa(milestone.GetNumber()),
		State:     "closed",
	}
	// check reponse for all issues
	issues, _, err := g.client.Issues.ListByRepo(g.ctx, g.owner, g.repo, &issueList)
	if err != nil {
		log.Fatal(err)
	}

	for _, issue := range issues {
		labels := createSet(issue)
		for _, section := range sections {
			if section.ContaintsLabels(labels) {
				section.Items = append(section.Items, gitHubItem{Issue: issue})
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
	_, _, err := g.client.Repositories.CreateRelease(g.ctx, g.owner, g.repo, &release)
	if err != nil {
		log.Fatal(err)
	}
}

// CreateRelease create release
func (g githubClientService) CloseVersion(version string) {
	milestone := g.findMilstone(version)
	if milestone.ClosedAt != nil {
		log.WithFields(log.Fields{
			"file":    version,
			"closeAt": milestone.GetClosedAt(),
		}).Warn("Version is already.")
		return
	}
	state := "closed"
	milestone.State = &state
	log.WithField("version", milestone.GetTitle()).Debug("Close milstone")
	_, _, err := g.client.Issues.EditMilestone(g.ctx, g.owner, g.repo, milestone.GetNumber(), milestone)
	if err != nil {
		log.Fatal(err)
	}
}

func (g githubClientService) findMilstone(version string) *github.Milestone {
	options := github.MilestoneListOptions{
		State: "open",
	}
	milestones, _, err := g.client.Issues.ListMilestones(g.ctx, g.owner, g.repo, &options)
	if err != nil {
		log.Fatal(err)
	}
	var milestone *github.Milestone
	for _, m := range milestones {
		if m.GetTitle() == version {
			milestone = m
			break
		}
	}
	if milestone == nil {
		log.WithField("version", version).Fatal("No open version found")
	}
	return milestone
}

// create set of labels
func createSet(g *github.Issue) map[string]bool {
	labels := make(map[string]bool)
	for _, label := range g.Labels {
		labels[label.GetName()] = true
	}
	return labels
}
