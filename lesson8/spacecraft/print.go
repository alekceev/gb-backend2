package main

import "fmt"

func Print(spacecraft map[string]interface{}) {
	name := spacecraft["name"]
	status := ""
	if spacecraft["status"] != nil {
		status = "- " + spacecraft["status"].(string)
	}
	registry := ""
	if spacecraft["registry"] != nil {
		registry = "- " + spacecraft["registry"].(string)
	}
	class := ""
	if spacecraft["spacecraftClass"] != nil {
		class = "- " +
			spacecraft["spacecraftClass"].(map[string]interface{})["name"].(string)
	}
	fmt.Println(name, registry, class, status)
}
