package polycode

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	TaskPending TaskStatus = iota
	TaskRunning
	TaskSuccess
	TaskFailed
	TaskCancelled
)

type TaskStatus int8

type StartAppRequest struct {
	AppName    string        `json:"appName"`
	AppPort    uint          `json:"appPort"`
	Services   []ServiceData `json:"services"`
	ApiHandler string        `json:"apiHandler"`
	Routes     []RouteData   `json:"routes"`
}

type ExecServiceRequest struct {
	Service      string      `json:"service"`
	TenantId     string      `json:"tenantId"`
	PartitionKey string      `json:"partitionKey"`
	Method       string      `json:"method"`
	Options      TaskOptions `json:"options"`
	Input        any         `json:"input"`
}

type ExecServiceExtendedRequest struct {
	EnvId              string             `json:"envId"`
	ExecServiceRequest ExecServiceRequest `json:"execServiceRequest"`
}

type ExecServiceResponse struct {
	IsAsync bool  `json:"isAsync"`
	Output  any   `json:"output"`
	IsError bool  `json:"isError"`
	Error   Error `json:"error"`
}

type ExecApiRequest struct {
	Controller string      `json:"controller"`
	Path       string      `json:"path"`
	Options    TaskOptions `json:"options"`
	Request    ApiRequest  `json:"request"`
}

type ExecApiExtendedRequest struct {
	EnvId          string         `json:"envId"`
	ExecApiRequest ExecApiRequest `json:"execApiRequest"`
}

type ExecApiResponse struct {
	IsAsync  bool        `json:"isAsync"`
	Response ApiResponse `json:"response"`
	IsError  bool        `json:"isError"`
	Error    Error       `json:"error"`
}

// PutRequest represents the JSON structure for put operations
type PutRequest struct {
	Action     string                 `json:"action"`
	Collection string                 `json:"collection"`
	Key        string                 `json:"key"`
	Item       map[string]interface{} `json:"item"`
}

// QueryRequest represents the JSON structure for query operations
type QueryRequest struct {
	Collection string        `json:"collection"`
	Key        string        `json:"key"`
	Filter     string        `json:"filter"`
	Args       []interface{} `json:"args"`
	Limit      int           `json:"limit"`
}

// QueryExtendedRequest represents the JSON structure for query operations
type QueryExtendedRequest struct {
	EnvId        string       `json:"envId"`
	TenantId     string       `json:"tenantId"`
	PartitionKey string       `json:"partitionKey"`
	ServiceName  string       `json:"serviceName"`
	QueryRequest QueryRequest `json:"queryRequest"`
}

// GetFileRequest represents the JSON structure for get file operations
type GetFileRequest struct {
	Key string `json:"key"`
}

type GetFileExtendedRequest struct {
	EnvId          string         `json:"envId"`
	TenantId       string         `json:"tenantId"`
	PartitionKey   string         `json:"partitionKey"`
	ServiceName    string         `json:"serviceName"`
	GetFileRequest GetFileRequest `json:"getFileRequest"`
}

// GetFileResponse represents the JSON structure for get file response
type GetFileResponse struct {
	Content string `json:"content"`
}

// PutFileRequest represents the JSON structure for put file operations
type PutFileRequest struct {
	Key     string `json:"key"`
	Content string `json:"content"`
}

// ServiceClient is a reusable client for calling the service API
type ServiceClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewServiceClient creates a new ServiceClient with a reusable HTTP client
func NewServiceClient(baseURL string) *ServiceClient {
	return &ServiceClient{
		httpClient: &http.Client{
			Timeout: time.Second * 30, // Set a reasonable timeout for HTTP requests
		},
		baseURL: baseURL,
	}
}

// StartApp starts the app
func (sc *ServiceClient) StartApp(req StartAppRequest) error {
	return executeApiWithoutResponse(sc.httpClient, sc.baseURL, "", "v1/system/app/start", req)
}

// ExecService executes a service with the given request
func (sc *ServiceClient) ExecService(sessionId string, req ExecServiceRequest) (ExecServiceResponse, error) {
	var res ExecServiceResponse
	err := executeApiWithResponse(sc.httpClient, sc.baseURL, sessionId, "v1/context/service/exec", req, &res)
	if err != nil {
		fmt.Printf("exec service error %s", err.Error())
		return ExecServiceResponse{}, err
	}

	if res.IsAsync {
		panic(ErrTaskStopped)
	}
	return res, nil
}

// ExecServiceExtended executes a service with the given request
func (sc *ServiceClient) ExecServiceExtended(sessionId string, req ExecServiceExtendedRequest) (ExecServiceResponse, error) {
	var res ExecServiceResponse
	err := executeApiWithResponse(sc.httpClient, sc.baseURL, sessionId, "v1/extended/context/service/exec", req, &res)
	if err != nil {
		fmt.Printf("exec service extended error %s", err.Error())
		return ExecServiceResponse{}, err
	}

	if res.IsAsync {
		panic(ErrTaskStopped)
	}
	return res, nil
}

func (sc *ServiceClient) ExecApi(sessionId string, req ExecApiRequest) (ExecApiResponse, error) {
	var res ExecApiResponse
	err := executeApiWithResponse(sc.httpClient, sc.baseURL, sessionId, "v1/context/api/exec", req, &res)
	if err != nil {
		fmt.Printf("exec api error %s", err.Error())
		return ExecApiResponse{}, err
	}

	if res.IsAsync {
		panic(ErrTaskStopped)
	}
	return res, nil
}

func (sc *ServiceClient) ExecApiExtended(sessionId string, req ExecApiExtendedRequest) (ExecApiResponse, error) {
	var res ExecApiResponse
	err := executeApiWithResponse(sc.httpClient, sc.baseURL, sessionId, "v1/extended/context/api/exec", req, &res)
	if err != nil {
		fmt.Printf("exec api extended error %s", err.Error())
		return ExecApiResponse{}, err
	}

	if res.IsAsync {
		panic(ErrTaskStopped)
	}
	return res, nil
}

// GetItem gets an item from the database
func (sc *ServiceClient) GetItem(sessionId string, req QueryRequest) (map[string]interface{}, error) {
	var res map[string]interface{}
	err := executeApiWithResponse(sc.httpClient, sc.baseURL, sessionId, "v1/context/db/get", req, &res)
	if err != nil {
		fmt.Printf("get item error %s", err.Error())
		return nil, err
	}
	return res, nil
}

// GetItemExtended gets an item from the database
func (sc *ServiceClient) GetItemExtended(sessionId string, req QueryExtendedRequest) (map[string]interface{}, error) {
	var res map[string]interface{}
	err := executeApiWithResponse(sc.httpClient, sc.baseURL, sessionId, "v1/extended/context/db/get", req, &res)
	if err != nil {
		fmt.Printf("get item extended error %s", err.Error())
		return nil, err
	}
	return res, nil
}

// QueryItems queries items from the database
func (sc *ServiceClient) QueryItems(sessionId string, req QueryRequest) ([]map[string]interface{}, error) {
	var res []map[string]interface{}
	err := executeApiWithResponse(sc.httpClient, sc.baseURL, sessionId, "v1/context/db/query", req, &res)
	if err != nil {
		fmt.Printf("query items error %s", err.Error())
		return nil, err
	}
	return res, nil
}

// QueryItemsExtended queries items from the database
func (sc *ServiceClient) QueryItemsExtended(sessionId string, req QueryExtendedRequest) ([]map[string]interface{}, error) {
	var res []map[string]interface{}
	err := executeApiWithResponse(sc.httpClient, sc.baseURL, sessionId, "v1/extended/context/db/query", req, &res)
	if err != nil {
		fmt.Printf("query items extended error %s", err.Error())
		return nil, err
	}
	return res, nil
}

// PutItem puts an item into the database
func (sc *ServiceClient) PutItem(sessionId string, req PutRequest) error {
	return executeApiWithoutResponse(sc.httpClient, sc.baseURL, sessionId, "v1/context/db/put", req)
}

// GetFile gets a file from the file store
func (sc *ServiceClient) GetFile(sessionId string, req GetFileRequest) (GetFileResponse, error) {
	var res GetFileResponse
	err := executeApiWithResponse(sc.httpClient, sc.baseURL, sessionId, "v1/context/file/get", req, &res)
	if err != nil {
		fmt.Printf("get file error %s", err.Error())
		return GetFileResponse{}, err
	}
	return res, nil
}

// GetFileExtended gets a file from the file store
func (sc *ServiceClient) GetFileExtended(sessionId string, req GetFileExtendedRequest) (GetFileResponse, error) {
	var res GetFileResponse
	err := executeApiWithResponse(sc.httpClient, sc.baseURL, sessionId, "v1/extended/context/file/get", req, &res)
	if err != nil {
		fmt.Printf("get file extended error %s", err.Error())
		return GetFileResponse{}, err
	}
	return res, nil
}

// PutFile puts a file into the file store
func (sc *ServiceClient) PutFile(sessionId string, req PutFileRequest) error {
	return executeApiWithoutResponse(sc.httpClient, sc.baseURL, sessionId, "v1/context/file/put", req)
}

func executeApiWithoutResponse(httpClient *http.Client, baseUrl string, sessionId string, path string, req any) error {
	log.Printf("client: exec api without response from %s with session id %s", path, sessionId)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", baseUrl, path), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-polycode-task-session-id", sessionId)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http error, status: %v", resp.Status)
	}

	return nil
}

func executeApiWithResponse[T any](httpClient *http.Client, baseUrl string, sessionId string, path string, req any, res *T) error {
	log.Printf("client: exec api with response from %s with session id %s\n", path, sessionId)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", baseUrl, path), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-polycode-task-session-id", sessionId)

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if res == nil {
		return errors.New("response is null")
	}

	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(res)
		if err != nil {
			return err
		}
		return nil
	} else {
		errorEvent := ErrorEvent{}
		err = json.NewDecoder(resp.Body).Decode(&errorEvent)
		if err != nil {
			return err
		}
		return errorEvent.Error
	}
}
