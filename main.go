package main

import (
	"log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"importer"
	"os"
)

func main() {
	loadYaml("config.yml", &importer.AppConfig)
	loadYaml("emotions.yml", &importer.AppEmotions)

	if importer.GetNeo4jConnection() == nil {
		os.Exit(1)
	}

	createConstraints()

	importer.ImportData()
}

func loadYaml(file string, container interface{}) {
	yamlData, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatalf("Cannot load file %s: %s", file, err)
	}

	err = yaml.Unmarshal(yamlData, container)

	if err != nil {
		log.Fatalf("Cannot parse YAML: %s", err)
	}
}

func createConstraints() {
	constrains := map[string]string{
		"Track": "id",
		"Tag": "name",
		"Artist": "id",
		"Album": "id",
		"Emotion": "name",
	}

	for label, id := range constrains {
		_, err := importer.GetNeo4jConnection().CreateUniqueConstraint(label, id)

		if err != nil {
			log.Printf("Cannot create constraint %s (%s): %s", label, id, err)
		}
	}
}