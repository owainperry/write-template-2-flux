package cmd

import (
	"crypto/rand"
	"encoding/hex"
	
	"github.com/spf13/cobra"
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		//os.Setenv("ROLENAME", "abc123")
		//os.Setenv()

		templatefolder, _ := cmd.Flags().GetString("templatefolder")
		emailAddress, _ := cmd.Flags().GetString("email")
		emailName, _ := cmd.Flags().GetString("username")
		gitBranch , _ := cmd.Flags().GetString("branch")
		url , _ := cmd.Flags().GetString("fluxuri")
		pushRetries, _ := cmd.Flags().GetInt("pushretry")
		log.Info(templatefolder)
		repoPath := tempName("tmp-", "")
		log.Infof("repo path: %s", repoPath)
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			log.Error("you need to set GITHUB_TOKEN envrionment variable , this is the only supported auth at this time")
			os.Exit(1)
		}


		auth := http.BasicAuth{}
		if token != "" {
			auth = http.BasicAuth{
				Username: "user",
				Password: token,
			}
		}

		tt := NewController(url, repoPath, auth, emailName, emailAddress, pushRetries, gitBranch)

		RunIt(tt, templatefolder)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("templatefolder", "t", "", "root folder containing templates")
	runCmd.MarkFlagRequired("templatefolder")
	runCmd.Flags().StringP("fluxuri", "f", "", "github url for flux repository")
	runCmd.MarkFlagRequired("fluxuri")
	runCmd.Flags().StringP("email", "e", "write-template-2-flux@example.com", "email address for git commit")
	runCmd.Flags().StringP("username", "u", "write-template-2-flux", "username for git commit")
	runCmd.Flags().StringP("branch", "b", "main", "branch to commit too")
	runCmd.Flags().IntP("pushretry", "p", 10, "number of retries while pushing changes")

}

func tempName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join("/tmp", prefix+hex.EncodeToString(randBytes)+suffix)
}

func getEnv(name string, defaultValue string) string {

	rtn := os.Getenv(name)
	if rtn == "" {
		rtn = defaultValue
	}
	return rtn
}

func RunIt(tt Controller, templatefolder string) {
	log.SetOutput(os.Stdout)

	err := filepath.Walk(templatefolder, tt.RenderAndAddFiles)
	if err != nil {
		log.Error(err)
	}

	tt.Commit("Automated commit")
	tt.Push()

	if err != nil {
		os.Exit(1)
	}

}
