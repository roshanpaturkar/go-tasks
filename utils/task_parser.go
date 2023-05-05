package utils

import "time"

func UpdateTaskParser(taskUpdate map[string]interface{}, existingMetadata map[string]string) map[string]interface{} {
	allowedKeys := []string{"title", "completed", "metadata"}
	for key := range taskUpdate {
		validKey := false
		for _, allowedKey := range allowedKeys {
			if key == allowedKey {
				validKey = true
			}
		}
		if !validKey {
			delete(taskUpdate, key)
		}
	}
	
	if _, ok := taskUpdate["metadata"]; ok {
		for key, value := range taskUpdate["metadata"].(map[string]interface{}) {
			existingMetadata[key] = value.(string)
		}
		taskUpdate["metadata"] = existingMetadata
	}

	taskUpdate["updated_at"] = time.Now().Unix()
	return taskUpdate
}