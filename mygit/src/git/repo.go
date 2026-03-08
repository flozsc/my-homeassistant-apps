package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var repoStorage string

func SetStorage(path string) {
	repoStorage = path
}

func GetStoragePath() string {
	return repoStorage
}

type Commit struct {
	SHA      string
	Message  string
	Author   string
	Email    string
	Date     string
	ParentSHAs []string
}

type Branch struct {
	Name         string
	IsRemote     bool
	IsCurrent    bool
	TrackingBranch string
}

type TreeEntry struct {
	Name     string
	Type     string // "tree" (directory) or "blob" (file)
	Size     int64
	Mode     string
}

type Contributor struct {
	Name    string
	Email   string
	Commits int
}

func GetRepoPath(name string) string {
	if !strings.HasSuffix(name, ".git") {
		name = name + ".git"
	}
	return filepath.Join(repoStorage, name)
}

func RepoExists(name string) bool {
	_, err := os.Stat(GetRepoPath(name))
	return err == nil
}

func CreateBareRepo(name string) error {
	path := GetRepoPath(name)
	if RepoExists(name) {
		return fmt.Errorf("repository already exists")
	}
	cmd := exec.Command("git", "init", "--bare", path)
	return cmd.Run()
}

func DeleteRepo(name string) error {
	path := GetRepoPath(name)
	if !RepoExists(name) {
		return fmt.Errorf("repository not found")
	}
	return os.RemoveAll(path)
}

func GetCommits(ctx context.Context, repoName, branch string, limit, offset int) ([]Commit, error) {
	path := GetRepoPath(repoName)
	if !RepoExists(repoName) {
		return nil, fmt.Errorf("repository not found")
	}

	ref := branch
	if ref == "" {
		ref = "HEAD"
	}

	cmd := exec.CommandContext(ctx, "git", "-C", path, "log",
		"--format=%H|%s|%an|%ae|%aI|%P",
		"-n", strconv.Itoa(limit),
		"--skip", strconv.Itoa(offset),
		ref,
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get commits: %v", err)
	}

	var commits []Commit
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 6)
		if len(parts) < 5 {
			continue
		}
		commit := Commit{
			SHA:     parts[0],
			Message: parts[1],
			Author:  parts[2],
			Email:   parts[3],
			Date:    parts[4],
		}
		if len(parts) > 5 && parts[5] != "" {
			commit.ParentSHAs = strings.Split(parts[5], " ")
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func GetBranches(ctx context.Context, repoName string) ([]string, error) {
	path := GetRepoPath(repoName)
	if !RepoExists(repoName) {
		return nil, fmt.Errorf("repository not found")
	}

	cmd := exec.CommandContext(ctx, "git", "-C", path, "branch", "-a", "--format=%(refname:short)")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get branches: %v", err)
	}

	var branches []string
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "HEAD") {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

func GetDefaultBranch(repoName string) (string, error) {
	path := GetRepoPath(repoName)
	if !RepoExists(repoName) {
		return "", fmt.Errorf("repository not found")
	}

	// Try to get default branch from git config
	cmd := exec.Command("git", "-C", path, "config", "--get", "init.defaultbranch")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		branch := strings.TrimSpace(out.String())
		if branch != "" {
			return branch, nil
		}
	}

	// Default to main or master
	for _, branch := range []string{"main", "master"} {
		cmd := exec.Command("git", "-C", path, "rev-parse", "--verify", branch)
		if cmd.Run() == nil {
			return branch, nil
		}
	}

	return "main", nil
}

func GetTree(ctx context.Context, repoName, ref, path string) ([]TreeEntry, error) {
	repoPath := GetRepoPath(repoName)
	if !RepoExists(repoName) {
		return nil, fmt.Errorf("repository not found")
	}

	var cmd *exec.Cmd
	var stderr bytes.Buffer
	if path == "" {
		cmd = exec.CommandContext(ctx, "git", "-C", repoPath, "ls-tree", ref)
	} else {
		cmd = exec.CommandContext(ctx, "git", "-C", repoPath, "ls-tree", ref, path)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get tree: %v: %s", err, stderr.String())
	}

	var entries []TreeEntry
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	re := regexp.MustCompile(`^(\d+)\s+(\w+)\s+(\S+)\s*(.*)$`)

	for _, line := range lines {
		if line == "" {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		entryType := matches[2]
		fullName := matches[3]
		name := matches[4]

		if path != "" {
			name = strings.TrimPrefix(fullName, path+"/")
		}

		entries = append(entries, TreeEntry{
			Name: name,
			Type: entryType,
			Mode: matches[1],
		})
	}

	// Get file sizes for blobs
	for i := range entries {
		if entries[i].Type == "blob" {
			size, _ := getFileSize(repoPath, ref, path, entries[i].Name)
			entries[i].Size = size
		}
	}

	return entries, nil
}

func getFileSize(repoPath, ref, dir, name string) (int64, error) {
	var filePath string
	if dir == "" {
		filePath = name
	} else {
		filePath = dir + "/" + name
	}

	cmd := exec.Command("git", "-C", repoPath, "ls-files", "-s", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 0, nil
	}

	// Format: <mode> <object> <stage> <filename>
	parts := strings.Fields(out.String())
	if len(parts) > 1 {
		cmd = exec.Command("git", "-C", repoPath, "cat-file", "-s", parts[1])
		var sizeOut bytes.Buffer
		cmd.Stdout = &sizeOut
		if err := cmd.Run(); err == nil {
			return strconv.ParseInt(strings.TrimSpace(sizeOut.String()), 10, 64)
		}
	}
	return 0, nil
}

func GetBlob(repoName, ref, path string) (string, error) {
	repoPath := GetRepoPath(repoName)
	if !RepoExists(repoName) {
		return "", fmt.Errorf("repository not found")
	}

	cmd := exec.Command("git", "-C", repoPath, "show", ref+":"+path)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get blob: %v", err)
	}

	return out.String(), nil
}

func GetBlobRaw(repoName, ref, path string) ([]byte, error) {
	repoPath := GetRepoPath(repoName)
	if !RepoExists(repoName) {
		return nil, fmt.Errorf("repository not found")
	}

	cmd := exec.Command("git", "-C", repoPath, "cat-file", "-p", ref+":"+path)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get blob: %v", err)
	}

	return out.Bytes(), nil
}

func GetContributors(repoName string) ([]Contributor, error) {
	path := GetRepoPath(repoName)
	if !RepoExists(repoName) {
		return nil, fmt.Errorf("repository not found")
	}

	cmd := exec.Command("git", "-C", path, "shortlog", "-sne", "--all")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to get contributors: %v", err)
	}

	var contributors []Contributor
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: <count>\t<name> <email>
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) < 2 {
			continue
		}

		count, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			continue
		}

		// Extract email from <>
		email := ""
		name := parts[1]
		if idx := strings.Index(name, "<"); idx >= 0 {
			endIdx := strings.Index(name, ">")
			if endIdx > idx {
				email = name[idx+1 : endIdx]
				name = strings.TrimSpace(name[:idx])
			}
		}

		contributors = append(contributors, Contributor{
			Name:    name,
			Email:   email,
			Commits: count,
		})
	}

	return contributors, nil
}

func GetLatestCommit(repoName, ref string) (string, error) {
	path := GetRepoPath(repoName)
	if !RepoExists(repoName) {
		return "", fmt.Errorf("repository not found")
	}

	cmd := exec.Command("git", "-C", path, "rev-parse", "--short", ref)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", nil
	}

	return strings.TrimSpace(out.String()), nil
}

func GetRepoSize(repoName string) (string, error) {
	path := GetRepoPath(repoName)
	if !RepoExists(repoName) {
		return "0 B", nil
	}

	cmd := exec.Command("du", "-sh", path)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "0 B", nil
	}

	return strings.TrimSpace(out.String()), nil
}
