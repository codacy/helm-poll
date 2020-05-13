# helm-poll

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/f7afd18c84b64493bc31a5db5bc53513)](https://app.codacy.com/gh/codacy/helm-poll?utm_source=github.com&utm_medium=referral&utm_content=codacy/helm-poll&utm_campaign=Badge_Grade_Dashboard)
[![CircleCI](https://circleci.com/gh/codacy/helm-poll/tree/master.svg?style=svg)](https://circleci.com/gh/codacy/helm-poll/tree/master)

A Helm plugin to poll for a release status.

## Project Goals

As per the [Helm documentation](https://helm.sh/docs/helm/helm_status/), a release can have the following final states:

```go
"unknown", "deployed", "deleted", "superseded", "failed"
```

A final state means that the release does not have any ongoing operations, such as an upgrade.

Therefore, to avoid concurrent installation problems, we have created this plugin that polls for the state of a given release until Helm returns one of the final states mentioned above.

This means that once the release reaches one of these states, there is no ongoing installation and we are free to proceed.

## Getting Started

The following command will install this plugin with your local copy of Helm.

Choose the latest version from the releases and install the appropriate version for your OS:

### Linux

```sh
helm plugin install https://github.com/codacy/helm-poll/releases/download/latest/helm-poll-linux.tgz
```

### MacOS

```sh
helm plugin install https://github.com/codacy/helm-poll/releases/download/latest/helm-poll-macos.tgz
```

### Windows

```sh
helm plugin install https://github.com/codacy/helm-poll/releases/download/latest/helm-poll-windows.tgz
```

### Prerequisites

* [Helm](https://helm.sh/)
* [GoLang](https://golang.org/)

Please note that this plugin will poll the state of the releases in the cluster that is in your current kubeconfig context. Make sure you are pointing to the desired cluster before running the plugin.

### Build and run

To build the plugin binary, all you really need to do is:

```bash
go build .
```

You can then run the produced binary with

```bash
$ ./poll

Usage: helm-poll [--help] [-i value] [-r value] [-t value] [--verbose] [-w value] [parameters ...]
     --help     Help
 -i, --interval=value
                The polling interval in seconds (default: 5)
 -r, --release=value
                Release name to poll for.
 -t, --timeout=value
                The timeout in seconds (default: 300)
     --verbose  Run with debug messages on
 -w, --with-namespace=value
                Namespace where the release is installed. (default: "default")
```

Upon success, the plugin will return the JSON output corresponding to the release:

```bash
helm poll -r codacy-nightly -w codacy
```

```json
{
  "Name": "codacy-nightly",
  "Revision": 47,
  "Updated": "Thu Jan 30 13:41:56 2020",
  "Status": "deployed",
  "Chart": "codacy-0.5.0-NIGHTLY.30-01-2020",
  "AppVersion": "0.5.0-NIGHTLY.30-01-2020",
  "Namespace": "codacy-nightly"
}
```

Polling for a release that does not exist will return an empty release:

```json
{
  "Name": "",
  "Revision": 0,
  "Updated": "",
  "Status": "",
  "Chart": "",
  "AppVersion": "",
  "Namespace": ""
}
```

Because the output is JSON, you can pipe it through `jq` for fancier operations.

## Running the tests

```bash
go test ./... -v
```

## Design

### Defaults

| Parameter          | Description                       | Default   | Required  |
| ------------------ | -------------------------------   | --------  | --------- |
| `--release`        | Name of the release to monitor    | `nil`     | True      |
| `--with-namespace` | Namespace where the release lives | `default` | False     |
| `--timeout`        | Polling timeout in seconds        | `300`     | False     |
| `--interval`       | Polling interval in seconds       | `5`       | False     |
| `--verbose`        | Enable debug messages             | `false`   | False     |

## What is Codacy

[Codacy](https://www.codacy.com/) is an Automated Code Review Tool that monitors your technical debt, helps you improve your code quality, teaches best practices to your developers, and helps you save time in Code Reviews.

### Among Codacy's features

- Identify new Static Analysis issues
- Commit and Pull Request Analysis with GitHub, BitBucket/Stash, GitLab (and also direct git repositories)
- Auto-comments on Commits and Pull Requests
- Integrations with Slack, HipChat, Jira, YouTrack
- Track issues in Code Style, Security, Error Proneness, Performance, Unused Code and other categories

Codacy also helps keep track of Code Coverage, Code Duplication, and Code Complexity.

Codacy supports PHP, Python, Ruby, Java, JavaScript, and Scala, among others.

Codacy is free for Open Source projects.

## License

helm-poll is available under the MIT license. See the [LICENSE](./LICENSE) file for more info.
