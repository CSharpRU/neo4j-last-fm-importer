package importer

import (
	"gopkg.in/jmcvetta/neoism.v1"
	"net/url"
	"fmt"
	"log"
	"github.com/imdario/mergo"
)

var neo4jConnection *neoism.Database

func GetNeo4jConnection() *neoism.Database {
	if neo4jConnection == nil {
		url := url.URL{
			Scheme: AppConfig.Neo4j.Scheme,
			User: url.UserPassword(AppConfig.Neo4j.Username, AppConfig.Neo4j.Password),
			Host: fmt.Sprintf("%s:%d", AppConfig.Neo4j.Host, AppConfig.Neo4j.Port),
			Path: "/db/data",
		}

		neo4jConnection, err := neoism.Connect(url.String())

		if err != nil {
			log.Printf("Cannot connect to neo4j: %s", err)
		}

		return neo4jConnection
	}

	return neo4jConnection
}

func GetOrCreateNode(label string, key string, props neoism.Props) (node *neoism.Node) {
	node, created, err := GetNeo4jConnection().GetOrCreateNode(label, key, props)

	if err != nil {
		log.Printf("Cannot create node: %s", err)
	}

	if created {
		err := node.AddLabel(label)

		if err != nil {
			log.Printf("Cannot add label: %s", err)
		}
	}

	return
}

func GetOrCreateRelationship(from *neoism.Node, to *neoism.Node, relType string, props neoism.Props) (relationship *neoism.Relationship) {
	relationships, err := from.Relationships(relType)

	if err == nil {
		for _, relationship := range relationships {
			endNode, err := relationship.End()

			if err != nil {
				continue
			}

			if endNode.Id() == to.Id() {
				newProps, err := relationship.Properties()

				if err != nil {
					return relationship
				}

				if err := mergo.Merge(&newProps, props); err != nil {
					relationship.SetProperties(newProps)
				}

				return relationship
			}
		}
	}

	relationship, err = from.Relate(relType, to.Id(), props)

	if err != nil {
		log.Printf("Cannot create relationship: %s", err)
	}

	return
}