package main

import (
	"bufio"
	"bytes"
	"encoding/json"
)

func Search(reader *bufio.Scanner, querytype string) {
	key := ReadText(reader, "Enter key")
	value := ReadText(reader, "Enter value")
	var buffer bytes.Buffer
	query := map[string]interface{}{
		"query": map[string]interface{}{
			querytype: map[string]interface{}{
				key: value,
			},
		},
	}
	json.NewEncoder(&buffer).Encode(query)

	response, _ := es.Search(es.Search.WithIndex("stsc"),
		es.Search.WithBody(&buffer))

	var result map[string]interface{}
	json.NewDecoder(response.Body).Decode(&result)
	
	for _, hit := range result["hits"].(map[string]interface{})["hits"].([]interface{}) {
		craft :=
			hit.(map[string]interface{})["_source"].(map[string]interface{})
		Print(craft)
	}
}
