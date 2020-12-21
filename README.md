# changelog

Changelog generator

[![License](https://img.shields.io/github/license/lorislab/changelog?style=for-the-badge&logo=apache)](https://www.apache.org/licenses/LICENSE-2.0)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/lorislab/changelog?logo=github&style=for-the-badge)](https://github.com/lorislab/changelog/releases/latest)

## Example
Write changelog to the console
```shell script
changelog generate --owner lorislab --repo release-notes --token xxxxxxxxxxxxxxxxxxxxxxx --version 2.0.0 --console
```
Create release and close version
```shell script
changelog generate --owner lorislab --repo release-notes --token xxxxxxxxxxxxxxxxxxxxxxx --version 2.0.0 --create-release --close-version
```
## Commands

```shell script
❯ changelog generate --help
Generate change for the release

Usage:
  changelog generate [flags]

Flags:
      --close-version    close version
      --console          write changelog to the console
      --create-release   create release and changelog
  -f, --file string      changelog definition (default "changelog.yaml")
  -h, --help             help for generate
  -w, --owner string     project owner (mandatory)
  -r, --repo string      repository name (mandatory)
  -t, --token string     access token
  -e, --version string   release version (mandatory)

Global Flags:
      --config string      config file (default is $HOME/.changelog.yaml)
  -v, --verbosity string   Log level (debug, info, warn, error, fatal, panic (default "info")
```