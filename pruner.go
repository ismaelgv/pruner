package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func difference(a []string, b []string) (diff []string) {
	m := make(map[string]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}

	return
}

func fetchRemote() error {
	command := exec.Command("git", "fetch")

	err := command.Run()
	if err != nil {
		return err
	}
	return nil
}

func pruneRemote() error {
	command := exec.Command("git", "remote", "prune", "origin")

	err := command.Run()
	if err != nil {
		return err
	}
	return nil
}

func getBranches(repository *git.Repository) ([]string, error) {

	var remoteBranches []string
	refs, err := repository.References()
	refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsRemote() {
			remoteBranch := strings.TrimPrefix(ref.Name().Short(), "origin/")
			remoteBranches = append(remoteBranches, remoteBranch)
		}
		return nil
	})

	// Get local branches from git
	branches, err := repository.Branches()
	if err != nil {
		return nil, err
	}

	var localBranches []string
	branches.ForEach(func(ref *plumbing.Reference) error {
		localBranches = append(localBranches, ref.Name().Short())
		return nil
	})

	return difference(localBranches, remoteBranches), nil
}

func run() error {
	repository, err := git.PlainOpen("./")
	if err != nil {
		return err
	}

	fetch := false
	fetchPrompt := &survey.Confirm{
		Message: "Do you want to fetch the remote repository?",
	}
	survey.AskOne(fetchPrompt, &fetch)
	if fetch {
		err := fetchRemote()
		if err != nil {
			return err
		}
	}

	prune := false
	prunePrompt := &survey.Confirm{
		Message: "Do you want to prune the remote repository?",
	}
	survey.AskOne(prunePrompt, &prune)
	if prune {
		err := pruneRemote()
		if err != nil {
			return err
		}
	}

	branches, err := getBranches(repository)
	if err != nil {
		return err
	}
	if len(branches) == 0 {
		fmt.Println("No local branches to prune.")
		return nil
	}

	branchesPrompt := &survey.MultiSelect{
		Message: "Select the local branches to delete:",
		Options: branches,
	}
	var selectedBranches []string
	survey.AskOne(branchesPrompt, &selectedBranches)

	for _, branch := range selectedBranches {
		fmt.Println("Remove branch:", branch)

		err := repository.DeleteBranch(branch)
		if err != nil {
			return err
		}

		refName := plumbing.NewBranchReferenceName(branch)
		err = repository.Storer.RemoveReference(refName)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
