package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"server/config"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/robertkrimen/otto"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

const (
	// URL -
	URL = "https://vhpi.5000.gov.tw/"
)

var (
	// Weeks -
	Weeks = [4]string{"winNo1", "winNo2", "winNo3", "winNo4"}

	// c -  新建一個定時任務物件
	c = &cron.Cron{}
)

func init() {
	location, _ := time.LoadLocation("Asia/Taipei")
	c = cron.New(cron.WithLocation(location))
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		server()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func server() {
	c.AddFunc("30-59/5 10,15 * * *", func() {
		reptile()
		autoCommit()
	})

	c.Start()

	for {
	}

	//reptile()
	//autoCommit()
}

func reptile() {
	log.Println("爬蟲開始....")
	data, err := getScriptData()
	if err != nil {
		log.Println(err)
		return
	}

	setData(data)

	log.Println("爬蟲結束~~~~")
}

func getScriptData() (string, error) {
	doc, err := goquery.NewDocument(URL)

	if err != nil {
		return "", err
	}

	s := doc.Find("script")

	script := s.Nodes[2].FirstChild.Data

	return script, nil
}

func setData(data string) {
	vm := otto.New()
	vm.Run(data)

	for index, v := range Weeks {
		if value, err := vm.Get(v); err == nil {
			data, _ := value.Export()

			for i, v := range config.Default {
				val, ok := data.(map[string]interface{})[v.ID]
				if ok {
					switch index {
					case 0:
						config.Default[i].Win.W1 = val.([]string)
					case 1:
						config.Default[i].Win.W2 = val.([]string)
					case 2:
						config.Default[i].Win.W3 = val.([]string)
					case 3:
						config.Default[i].Win.W4 = val.([]string)
					}
				}
			}
		}
	}

	buildJson()
}

func buildJson() {
	file, _ := json.Marshal(config.Default)
	str := "const EXTRA_BONUS_LIST = " + string(file)
	err := ioutil.WriteFile("code.js", []byte(str), 0644)

	if err != nil {
		log.Println("write file is failed: ", err)
	}
}

// auto git commit
func autoCommit() {
	gitPullCmd := exec.Command("git", "pull", "--all")
	gitAddCmd := exec.Command("git", "add", ".")
	gitCommitCmd := exec.Command("git", "commit", "-am", fmt.Sprintf("update code.js at %v", time.Now().Format(time.RFC3339)))
	gitPushCmd := exec.Command("git", "push", "origin", "master")

	log.Println("git pull --all")
	if _, err := gitPullCmd.CombinedOutput(); err != nil {
		log.Println("git pull --all failed")
	}

	log.Println("git add .")
	if _, err := gitAddCmd.CombinedOutput(); err != nil {
		log.Println("git add . failed")
	}

	log.Println("git commit")
	if _, err := gitCommitCmd.CombinedOutput(); err != nil {
		log.Println("git commit failed")
	}

	log.Println("git push")
	if _, err := gitPushCmd.CombinedOutput(); err != nil {
		log.Println("git push failed")
	}
}
