package db

import (
	"log"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/review/backend/internal/config"
)

var ES *elasticsearch.Client

func ConnectElasticsearch(cfg *config.Config) {
	esConfig := elasticsearch.Config{
		Addresses: []string{
			cfg.ESUrl,
		},
	}

	es, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Error: %s", res.String())
	}

	log.Println("Connected to Elasticsearch successfully")
	ES = es
}
