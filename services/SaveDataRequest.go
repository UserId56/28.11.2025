package services

import (
	"encoding/json"
	"os"

	"25.11.2025/models"
)

func SaveDataRequest(taskList map[int]models.Task, fullPath string) error {
	jsonData, err := json.Marshal(taskList)
	if err != nil {
		return err
	}
	return os.WriteFile(fullPath, jsonData, 0644)
}

func ReadDataRequest(fileName string) (map[int]models.Task, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var taskList map[int]models.Task
	err = json.Unmarshal(data, &taskList)
	if err != nil {
		return nil, err
	}
	return taskList, nil
}
