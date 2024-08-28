package client

import (
	"a21hc3NpZ25tZW50/config"
	"a21hc3NpZ25tZW50/model"
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

type TaskClient interface {
	TaskList(token string) ([]*model.Task, error)
	AddTask(token string, task model.Task) (respCode int, err error)
	UpdateTask(token string, task model.Task) (respCode int, err error)
	DeleteTask(token string, id int) (respCode int, err error)
}

type taskClient struct {
}

func NewTaskClient() *taskClient {
	return &taskClient{}
}

func (t *taskClient) TaskList(token string) ([]*model.Task, error) {
	client, err := GetClientWithCookie(token)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", config.SetUrl("/api/v1/task/list"), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("status code not 200")
	}

	var tasks []*model.Task
	err = json.Unmarshal(b, &tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (t *taskClient) AddTask(token string, task model.Task) (respCode int, err error) {
	client, err := GetClientWithCookie(token)
	if err != nil {
		return -1, err
	}

	datajson := map[string]interface{}{
		"title":       task.Title,
		"deadline":    task.Deadline,
		"priority":    task.Priority,
		"status":      task.Status,
		"category_id": task.CategoryID,
		"user_id":     task.UserID,
	}

	data, err := json.Marshal(datajson)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("POST", config.SetUrl("/api/v1/task/add"), bytes.NewBuffer(data))
	if err != nil {
		return -1, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return -1, errors.New("status code not 200")
	}

	return resp.StatusCode, nil
}

func (t *taskClient) UpdateTask(token string, task model.Task) (respCode int, err error) {
	client, err := GetClientWithCookie(token)
	if err != nil {
		return -1, err
	}

	datajson := map[string]interface{}{
		"title":       task.Title,
		"deadline":    task.Deadline,
		"priority":    task.Priority,
		"status":      task.Status,
		"category_id": task.CategoryID,
		"user_id":     task.UserID,
	}

	data, err := json.Marshal(datajson)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("PUT", config.SetUrl("/api/v1/task/update/"+strconv.Itoa(task.ID)), bytes.NewBuffer(data))
	if err != nil {
		return -1, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return -1, errors.New("status code not 200")
	}

	return resp.StatusCode, nil
}

func (t *taskClient) DeleteTask(token string, id int) (respCode int, err error) {
	client, err := GetClientWithCookie(token)
	if err != nil {
		return -1, err
	}

	req, err := http.NewRequest("DELETE", config.SetUrl("/api/v1/task/delete/"+strconv.Itoa(id)), nil)
	if err != nil {
		return -1, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return -1, errors.New("status code not 200")
	}

	return resp.StatusCode, nil
}
