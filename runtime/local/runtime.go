package local

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/CloudImpl-Inc/next-coder-sdk/client"
	"github.com/CloudImpl-Inc/next-coder-sdk/client/db"
	"github.com/CloudImpl-Inc/next-coder-sdk/polycode"
	"github.com/gin-gonic/gin"
	"io"
)

//var runtime *Runtime = nil

type Runtime struct {
	dbClient *db.Client
}

func (r *Runtime) GetRuntime() polycode.Runtime {
	return r
}

func (r *Runtime) InvokeWorkflow(workflowContext polycode.WorkflowContext, input polycode.TaskInput) (any, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Runtime) Name() string {
	return "local"
}

// Invoke Handler implementation
func (r *Runtime) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	fmt.Printf("data received to lambda function %s\n", string(payload))
	evt := client.Event{}

	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}

	runtimeContext := client.NewRuntimeContext(ctx, db.NewDatabase(r.dbClient, evt.Context.SessionId))
	ret, err := client.RunTask(runtimeContext, r, evt)
	if err != nil {
		return nil, err
	}

	return json.Marshal(ret)
}

func (r *Runtime) Start(params []any) error {
	r.dbClient = db.NewClient("http://localhost:6666")
	//runtime = r
	//start gin server
	gin := gin.Default()
	gin.POST("/invoke", r.OnRequest)
	gin.Run(":8080")
	return nil
}

func (r *Runtime) OnRequest(c *gin.Context) {

	//call runtime invoke function

	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "error reading request body",
		})
		return
	}

	ret, err := r.Invoke(c, bodyBytes)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "error invoking runtime",
		})
		return
	}

	c.JSON(200, gin.H{
		"response": string(ret),
	})
}
