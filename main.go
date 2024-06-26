package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/apenella/go-ansible/v2/pkg/execute"
	"github.com/apenella/go-ansible/v2/pkg/execute/measure"
	results "github.com/apenella/go-ansible/v2/pkg/execute/result/json"
	"github.com/apenella/go-ansible/v2/pkg/execute/stdoutcallback"
	"github.com/apenella/go-ansible/v2/pkg/playbook"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

type ProcessInfo struct {
	ID interface{} `json:"id"`
	Name interface{} `json:"name"`
	Job interface{} `json:"job"`
	Creator interface{} `json:"creator"`
}


func GetInfo(c *gin.Context) {

	var err error
	var res *results.AnsiblePlaybookJSONResults

	buff := new(bytes.Buffer)

	// 这个变量作为主机名
	targetHost , ok := c.GetQuery("targetHost")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"err": "targetHost"})
	}
	targetPort , ok := c.GetQuery("targetPort")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"err": "targetPort"})
	}
	//targetHost := "node"
	//targetPort := 80
	ansiblePlaybookOptions := &playbook.AnsiblePlaybookOptions{
		ExtraVars: map[string]interface{}{
			"target_host": fmt.Sprintf("%s", targetHost),
			"target_port": fmt.Sprintf("%v", targetPort),
		},
	}

	playbookCmd := playbook.NewAnsiblePlaybookCmd(
		playbook.WithPlaybooks("dpa.yml"),
		playbook.WithPlaybookOptions(ansiblePlaybookOptions),
	)

	//fmt.Println("Command: ", playbookCmd.String())

	exec := measure.NewExecutorTimeMeasurement(
		stdoutcallback.NewJSONStdoutCallbackExecute(
			execute.NewDefaultExecute(
				execute.WithCmd(playbookCmd),
				execute.WithErrorEnrich(playbook.NewAnsiblePlaybookErrorEnrich()),
				execute.WithWrite(io.Writer(buff)),
			),
		),
	)

	err = exec.Execute(context.TODO())
	if err != nil {
		fmt.Println(err.Error())
	}

	res, err = results.ParseJSONResultsStream(io.Reader(buff))
	if err != nil {
		panic(err)
	}

	var processInfo ProcessInfo

	for _, play := range res.Plays {
		for _, task := range play.Tasks {
			if task.Task.Name == "PrintProcessID" {
				//for host, result := range task.Hosts {
				for _, result := range task.Hosts {
					processInfo.ID = result.Msg
				}
			}

			if task.Task.Name == "PrintProcessName" {
				for _, result := range task.Hosts {
					processInfo.Name = result.Msg
				}
			}

			if task.Task.Name == "PrintJobName" {
				for _, result := range task.Hosts {
					processInfo.Job = result.Msg
				}
			}

			if task.Task.Name == "PrintUsername" {
				for _, result := range task.Hosts {
					processInfo.Creator = result.Msg
				}
			}
		}
	}
	c.JSON(http.StatusOK, processInfo)
}



func main() {
	r := gin.Default()
	r.GET("/", GetInfo)
	r.Run("0.0.0.0:8080")
}
