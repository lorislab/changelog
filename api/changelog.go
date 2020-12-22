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
	FindVersionIssues(version string, groups []*Section)
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

// Section list of items for the label
type Section struct {
	Config ConfigSection
	Items  []Item
}

// GetTitle get group title
func (c *Section) GetTitle() string {
	return c.Config.Title
}

// ContaintsLabels check if section containts one of the labels
func (c *Section) ContaintsLabels(labels map[string]bool) bool {
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
	Version  string
	File     string
	Sections []*Section
	Client   ClientService
	Body     string
	SemVer   *semver.Version
	Config   *Config
}

// ConfigSection changelog configuration section
type ConfigSection struct {
	Title  string   `yaml:"title"`
	Labels []string `yaml:"labels"`
}

// Config changelog configuration
type Config struct {
	Template string          `yaml:"template"`
	Sections []ConfigSection `yaml:"sections"`
}

// Init initialize changelog
func (c *Changelog) Init() {
	c.Config = loadConfig(c.File)

	c.Sections = []*Section{}
	for _, section := range c.Config.Sections {
		c.Sections = append(c.Sections, &Section{Config: section, Items: []Item{}})
	}

	semVer, err := semver.NewVersion(c.Version)
	if err != nil {
		log.WithFields(log.Fields{
			"version": c.Version,
		}).Fatal(err)
	}
	c.SemVer = semVer
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
	c.Client.FindVersionIssues(c.Version, c.Sections)
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
	// check default Sections
	if len(config.Sections) == 0 {
		config.Sections = []ConfigSection{{Title: "Complete changelog", Labels: []string{"bug", "enhancement"}}}
	}
	return config
}

var (
	defaultTemplate = `{{ range $section := .Sections }}### {{ $section.GetTitle }}{{ range $item := $section.Items }}
* [#{{ $item.GetID }}]({{ $item.GetURL }}) - {{ $item.GetTitle }}{{ end }}
{{ end }}
`
)
