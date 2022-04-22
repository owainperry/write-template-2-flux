package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"strings"

	"github.com/a8m/envsubst"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	log "github.com/sirupsen/logrus"
	
)

type Controller struct {
	RepoPath     string
	WorkTree     *git.Worktree
	GitRepo      *git.Repository
	RepoUrl      string
	gitName      string
	gitEmail     string
	PushRetries  int
	gitBranch    string
	auth         http.BasicAuth
	templatePath string
}

func NewController(repoUrl string, repoPath string, auth http.BasicAuth, gitName string, gitEmail string, pushRetries int, gitBranch string, templatePath string) Controller {
	t := Controller{
		RepoPath:     repoPath,
		RepoUrl:      repoUrl,
		PushRetries:  pushRetries,
		gitName:      gitName,
		gitEmail:     gitEmail,
		gitBranch:    gitBranch,
		auth:         auth,
		templatePath: templatePath,
	}

	var err error
	t.GitRepo, err = git.PlainClone(t.RepoPath, false, &git.CloneOptions{
		Auth:     &t.auth,
		URL:      t.RepoUrl,
		Progress: nil,
	})
	t.ExitOnError(err)

	t.WorkTree, err = t.GitRepo.Worktree()
	t.ExitOnError(err)
	// err = t.WorkTree.Checkout(&git.CheckoutOptions{
	// 	Branch:  plumbing.ReferenceName(t.gitBranch),
	// })
	// t.ExitOnError(err)
	return t
}

func (t Controller) ExitOnError(err error) {
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func (t Controller) Commit(message string) {
	commit, err := t.WorkTree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  t.gitName,
			Email: t.gitEmail,
			When:  time.Now(),
		},
	})
	t.ExitOnError(err)
	log.Infof("Commit %s", commit.String())
}

func (t Controller) Push() {

	remoteName := "origin"
	success := true
	err := t.GitRepo.Push(&git.PushOptions{Auth: &t.auth})
	if err != nil {
		log.Error("Failed to push %s", err)
		for i := 1; i <= t.PushRetries; i++ {
			success = false
			time.Sleep(1 * time.Second)
			err = t.WorkTree.Pull(&git.PullOptions{RemoteName: remoteName, Auth: &t.auth})
			if err != nil {
				log.Errorf("Failed to pull from origin %s %s", remoteName, err)
			}
			err = t.GitRepo.Push(&git.PushOptions{Auth: &t.auth})
			if err != nil {
				log.Errorf("Failed to push %s %d", err, i)
			} else {
				success = true
				break
			}
		}
	}
	if success != true {
		t.ExitOnError(errors.New("Failed to push to origin within the number of retries"))
	}
	log.Info("Pushed all files successfully")
}

func (t Controller) RenderAndAddFiles(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	corePath := strings.ReplaceAll(path, t.templatePath, "")
	outputPath := filepath.Join(t.RepoPath, corePath)

	if corePath == "" {
		return nil
	}
	if info.IsDir() {
		err = os.MkdirAll(outputPath, os.ModePerm)
		if err != nil {
			log.Errorf(err.Error())
		}
		return nil
	}

	log.Info("processing:", path)
	buf, err := envsubst.ReadFile(path)
	if err != nil {
		log.Errorf(err.Error())
	}

	err = os.WriteFile(outputPath, buf, 0644)
	if err != nil {
		log.Errorf(err.Error())
	}
	log.Infof("Adding file %s",outputPath)
	err = t.WorkTree.AddGlob(".")
	if err != nil {
		log.Errorf(err.Error())
	}

	status, err := t.WorkTree.Status()
	if err != nil {
		log.Errorf(err.Error())
	}
	log.Info(status)
	return nil
}
