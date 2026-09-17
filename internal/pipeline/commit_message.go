package pipeline

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type commitMessageData struct {
	ReleaseVersion string
	NextVersion    string
	Project        string
	Package        string
}

func defaultReleaseCommitMessage(packageName, releaseVersion string) string {
	return "Release " + releaseCommitLabel(packageName, releaseVersion)
}

func releaseCommitLabel(packageName, releaseVersion string) string {
	if packageName != "" {
		return packageName + " " + releaseVersion
	}
	return releaseVersion
}

func defaultNextCommitMessage(packageName string) string {
	if packageName != "" {
		return "Prepare next development iteration for " + packageName
	}
	return "Prepare next development iteration"
}

func renderCommitMessage(kind, source, fallback string, data commitMessageData) (string, error) {
	if source == "" {
		return fallback, nil
	}
	tmpl, err := template.New(kind).Option("missingkey=error").Parse(source)
	if err != nil {
		return "", fmt.Errorf("parsing %s commit message template: %w", kind, err)
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("executing %s commit message template: %w", kind, err)
	}
	if strings.TrimSpace(out.String()) == "" {
		return "", fmt.Errorf("%s commit message template produced an empty message", kind)
	}
	return out.String(), nil
}

func renderBatchCommitMessage(kind string, results []*PackageResult, nextVersions []string) (string, error) {
	var packages, releaseVersions, projects []string
	source := ""
	for _, result := range results {
		packages = append(packages, result.Package.Name)
		releaseVersions = append(releaseVersions, result.NewVersion.String())
		project := result.Config.Name
		if project == "" {
			project = result.Config.Project
		}
		projects = append(projects, project)
		candidate := result.Config.Commit.Release
		if kind == "next" {
			candidate = result.Config.Commit.Next
		}
		if candidate != "" && source != "" && candidate != source {
			return "", fmt.Errorf("all packages in a batched release must use the same %s commit message template", kind)
		}
		if candidate != "" {
			source = candidate
		}
	}

	fallback := "Release " + strings.Join(labels(packages, releaseVersions), ", ")
	if kind == "next" {
		fallback = defaultNextCommitMessage("")
	}
	return renderCommitMessage(kind, source, fallback, commitMessageData{
		ReleaseVersion: strings.Join(releaseVersions, ", "),
		NextVersion:    strings.Join(nextVersions, ", "),
		Project:        strings.Join(projects, ", "),
		Package:        strings.Join(packages, ", "),
	})
}

func labels(packages, versions []string) []string {
	result := make([]string, len(packages))
	for i := range packages {
		result[i] = releaseCommitLabel(packages[i], versions[i])
	}
	return result
}
