package api

import (
	"bytes"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// ClientService client service
type ClientService interface {
	// FindVersionIssues find issues for the release
	FindVersionIssues(version string, groups []*Group)
	// CreateRelease create release
	CreateRelease(version string, prerelease bool, output string)
	// CloseVersion close current version
	CloseVersion(version string)
}

// Item of the group of items
type Item interface {
	GetID() string
	GetURL() string
	GetTitle() string
}

// Group list of items for the label
type Group struct {
	Config ConfigGroup
	Items  []Item
}

// GetTitle get group title
func (c *Group) GetTitle() string {
	return c.Config.Title
}

// ContaintsLabels check if group containts one of the labels
func (c *Group) ContaintsLabels(labels map[string]bool) bool {
	for _, label := range c.Config.Labels {
		_, exists := labels[label]
		if exists {
			return true
		}
	}
	return false
}

// Changelog main object
type Changelog struct {
	Version string
	File    string
	Groups  []*Group
	Client  ClientService
	Body    string
	SemVer  semver.Version
	Config  *Config
}

// ConfigGroup changelog configuration group
type ConfigGroup struct {
	Title  string   `yaml:"title"`
	Labels []string `yaml:"labels"`
}

// Config changelog configuration
type Config struct {
	Template string        `yaml:"template"`
	Groups   []ConfigGroup `yaml:"groups"`
}

// Init initialize changelog
func (c *Changelog) Init() {
	c.Config = loadConfig(c.File)

	c.Groups = []*Group{}
	for _, group := range c.Config.Groups {
		c.Groups = append(c.Groups, &Group{Config: group, Items: []Item{}})
	}

	semVer, err := semver.NewVersion(c.Version)
	if err != nil {
		log.Fatal(err)
	}
	c.SemVer = *semVer
}

// GenerateBody generate the body of the release
func (c *Changelog) GenerateBody() {
	bodyTemplate := defaultTemplate
	if len(c.Config.Template) > 0 {
		bodyTemplate = c.Config.Template
	}
	template, err := template.New("changelog").Parse(bodyTemplate)
	if err != nil {
		log.Panic(err)
	}
	var tpl bytes.Buffer
	err = template.Execute(&tpl, c)
	if err != nil {
		log.Panic(err)
	}
	c.Body = tpl.String()
}

// FindVersionIssues find issues for the version
func (c *Changelog) FindVersionIssues() {
	log.WithField("version", c.Version).Info("Find issues for version")
	c.Client.FindVersionIssues(c.Version, c.Groups)
}

// IsPrerelease returns true is the version is pre-release
func (c *Changelog) IsPrerelease() bool {
	return len(c.SemVer.Prerelease()) > 0
}

// CreateRelease create release
func (c *Changelog) CreateRelease() {
	prerelease := c.IsPrerelease()
	log.WithFields(log.Fields{
		"version":    c.Version,
		"prerelease": prerelease,
	}).Info("Create release for version")
	c.Client.CreateRelease(c.Version, prerelease, c.Body)
}

// CloseVersion close version
func (c *Changelog) CloseVersion() {
	log.WithField("version", c.Version).Infof("Close version")
	c.Client.CloseVersion(c.Version)
}

// LoadConfig load config
func loadConfig(file string) *Config {

	config := &Config{}
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.WithField("file", file).Debug("Configuration file does not exists.")
	} else {
		log.WithField("file", file).Debug("Load configuration from the file.")
		yamlFile, err := ioutil.ReadFile(file)
		if err != nil {
			log.Panic(err)
		}

		err = yaml.Unmarshal(yamlFile, config)
		if err != nil {
			log.Panic(err)
		}
	}
	// check default template value
	if len(config.Template) == 0 {
		config.Template = defaultTemplate
	}
	// check default groups
	if len(config.Groups) == 0 {
		config.Groups = []ConfigGroup{{Title: "Complete changelog", Labels: []string{"bug", "enhancement"}}}
	}
	return config
}

var (
	defaultTemplate = `{{ range $group := .Groups }}### {{ $group.GetTitle }}{{ range $item := $group.Items }}
* [#{{ $item.GetID }}]({{ $item.GetURL }}) - {{ $item.GetTitle }}{{ end }}
{{ end }}
`
)
