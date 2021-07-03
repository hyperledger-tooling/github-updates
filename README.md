# GITHUB UPDATES

The tool will assist in getting the PRs curated for the weekly developer
newsletter sent by Hyperledger. The tool will print list of all PRs from
Hyperledger and Hyperledger Labs into a html file that were raised in
last `X` days. The number of days is configurable, current default is set
to be 7. Similarly, it will pull all the releases created in last `X` days.

In addition to the PRs, the tool also polls for the issues with specific
tags and tracks releases on the input organization repositories. It is
customizable to enable any one or all three of these.

Note that the tool expects user to set the GitHub personal access token
with the read access. This is to avoid GitHub from blocking the machine due
to frequent API calls. The tool makes use of
[go-github](https://github.com/google/go-github/) for pulling the data.

## Dependencies

Option GitHub personal access token with read access (this is required
in case of frequent API calls, default rate that GitHub allow is 60).

Set the following environment variable before running the tool

```bash
export GITHUB_TOKEN=<YOUR LONG PERSONAL ACCESS TOKEN HERE>
```

The tool is written in Go version 1.15, you can also use the docker
container runtime engine to package and run it as a container.
Tool also comes with a `docker-compose` file to make it easy to run
the command with default configuration.

Tested `docker` and `docker-compose` versions

- Docker version 19.03.4
- docker-compose version 1.24.1

## Configuration

The tool can be configured to fetch the PRs from any organization via
the configuration as long as the input `GITHUB_TOKEN` has read access.
Pass the configuration file as below to the tool, find the comments
next to them here

```yaml

# Global configuration for all runs
global:
  # Organizations from where the PRs, Issues and Releases are to be listed. Each organization
  # has to be listed as a list element.
  organizations:
    - organization:
        name: "Hyperledger"
        github: "hyperledger"
    - organization:
        name: "Hyperledger Labs"
        github: "hyperledger-labs"
  scrape-duration-days: 7
  # Set this to true and specify input/output files
  external-template:
    enabled: false
    # Possible values "repository"
    template-for: "repository"

# Config for Issues
issue:
  # List tags which are to be matched and scraped. An issue is selected if at least one of the tags match
  issue-tags:
    - "good first issue"
    - "help wanted"
  # Issues created in the last N days to be listed
  created-history-days: 100
  # Report summary file
  summary-filename: "html/generated/issue-summary.html"
  # Should this report run?
  should-run: true
  # Data file for raw output
  data-file: "generated-data/issue-data.json"
  # Applicable if globally external-template is enabled
  external-template:
    # Input template file
    input: ""
    # Output file path, the generated file will with the repo name
    output: ""

# Config for Pull Requests
pull-requests:
  # Report summary file
  summary-filename: "html/generated/pr-summary.html"
  # Should this report run?
  should-run: true
  # Data file for raw output
  data-file: "generated-data/pr-data.json"
  # Applicable if globally external-template is enabled
  external-template:
    # Input template file
    input: ""
    # Output file path, the generated file will with the repo name
    output: ""

# Config for Releases
releases:
  # Report summary file
  summary-filename: "html/generated/release-summary.html"
  # Should this report run?
  should-run: true
  # Data file for raw output
  data-file: "generated-data/release-data.json"
  # Applicable if globally external-template is enabled
  external-template:
    # Input template file
    input: ""
    # Output file path, the generated file will with the repo name
    output: ""
```

## Environment

The tool accepts following environment variables in addition to
the configuration file.


```bash
# The html print file path for PRs
PR_SUMMARY_FILE_PATH
# The html print file path for releases
RELEASE_SUMMARY_FILE_PATH
# Issue summary html path
ISSUE_SUMMARY_FILE_PATH
# GitHub access token
GITHUB_TOKEN
# Configuration file path
CONFIG_FILE
```

## Development

As a developer if you would like to use container environment, run the following command

```bash
docker-compose -f build/docker-compose-build.yaml up github-updates-make
```

If you get an error for lint checks, run the following command

```bash
docker-compose -f build/docker-compose-build.yaml up github-updates-format
```

## Run

Prefer to run it as a container, run the following command

```bash
docker-compose -f deployments/docker-compose.yaml up
```

You may optionally build and run the tool as

```bash
make
./github-updates
```
