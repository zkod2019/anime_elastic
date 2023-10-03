package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	//	"github.com/google/uuid"

	api "github.com/elastic/go-elasticsearch/v8/esapi"
)

var elasticClient *elasticsearch8.Client

func main() {
	var err error
	elasticClient, err = elasticsearch8.NewClient(elasticsearch8.Config{
		Addresses: []string{"http://localhost:9200"},
	})

	if err != nil {
		panic(err)
	}

	router := gin.Default()
	api := router.Group("anime")
	api.POST("", func(ctx *gin.Context) {
		body := Anime{}
		err := ctx.BindJSON(&body)
		if err != nil {
			ctx.JSON(400, err)
			return
		}
		err = Create(body)
		if err != nil {
			fmt.Println(err)
			ctx.JSON(400, err)
			return
		}
		ctx.JSON(201, nil)

	})
	api.GET("", func(ctx *gin.Context) {
		list, err := Read()
		if err != nil {
			fmt.Println(err)
			ctx.JSON(400, err)
			return
		}
		ctx.JSON(200, list)
	})
	api.PUT("/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		body := Anime{}
		err := ctx.BindJSON(&body)
		if err != nil {
			ctx.JSON(400, err)
		}
		err = Update(id, body)
		if err != nil {
			fmt.Println(err)
			ctx.JSON(400, err)
			return
		}
		ctx.JSON(202, nil)
	})
	api.DELETE("/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		err := Delete(id)
		if err != nil {
			fmt.Println(err)
			ctx.JSON(400, err)
			return
		}
		ctx.JSON(204, nil)
	})
	api.POST("/search", func(ctx *gin.Context) {
		var elasticSearchQuery ElasticSearchQuery
		err := ctx.BindJSON(&elasticSearchQuery) // always need a pointer when binding/unmarshalling
		if err != nil {
			fmt.Println(err)
			ctx.JSON(400, err)
			return
		}
		list, err := Query(elasticSearchQuery.SearchTerm)
		if err != nil {
			fmt.Println(err)
			ctx.JSON(400, err)
			return
		}
		ctx.JSON(200, list)

	})
	router.Run(":8080")
}

func Create(cmd Anime) error {
	byts, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	req := api.IndexRequest{
		Index:      "anime",
		DocumentID: uuid.NewString(),
		Body:       bytes.NewReader(byts),
		Refresh:    "true",
	}
	_, err = req.Do(context.TODO(), elasticClient)
	if err != nil {
		return err
	}

	return nil
}

func Read() ([]Anime, error) { // read all docs
	response, err := elasticClient.Search(
		elasticClient.Search.WithContext(context.TODO()),
		elasticClient.Search.WithIndex("anime"),
		elasticClient.Search.WithBody(strings.NewReader(`{"query": {"match_all": {}}}`)),
		elasticClient.Search.WithTrackTotalHits(true),
		elasticClient.Search.WithPretty(),
	)
	if err != nil {
		return nil, err
	}
	var animeRes []Anime
	var elasticRes ElasticResponse
	err = json.NewDecoder(response.Body).Decode(&elasticRes)
	if err != nil {
		return nil, err
	}
	for i := range elasticRes.Hits.Hits {
		fmt.Println(elasticRes.Hits.Hits[i].ID)
		animeRes = append(animeRes, elasticRes.Hits.Hits[i].Source)
	}

	return animeRes, nil
}

func Update(id string, cmd Anime) error {
	fmt.Println("ID IS HERE:" + id)
	byts, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	// client.Update("my_index", "id", strings.NewReader(`{doc: { language: "Go" }}`))
	// elasticClient.Update(
	// 	"anime",
	// 	id,
	// 	bytes.NewReader(byts),
	// )
	req := api.UpdateRequest{
		Index:      "anime",
		DocumentID: id,
		Body:       bytes.NewReader([]byte(fmt.Sprintf(`{"doc":%s}`, byts))), // need to wrap the doc with the updates
		//Refresh:    "true",
	}

	_, err = req.Do(context.TODO(), elasticClient)
	if err != nil {
		return err
	}

	return nil
}

func Delete(id string) error {
	req := api.DeleteRequest{
		Index:      "anime",
		DocumentID: id,
	}
	_, err := req.Do(context.TODO(), elasticClient)
	if err != nil {
		return err
	}

	return nil
}

func Query(query string) ([]Anime, error) { // search for specific doc
	query = fmt.Sprintf(`
		{
		  "track_total_hits": false,
		  "sort": [
			{
			  "_doc": {
				"order": "desc",
				"unmapped_type": "boolean"
			  }
			}
		  ],
		  "fields": [
			{
			  "field": "*",
			  "include_unmapped": "true"
			},
			{
			  "field": "event.createdDateTime",
			  "format": "strict_date_optional_time"
			}
		  ],
		  "size": 500,
		  "version": true,
		  "script_fields": {},
		  "stored_fields": [
			"*"
		  ],
		  "runtime_mappings": {},
		  "_source": true,
		  "query": {
			"bool": {
			  "must": [],
			  "filter": [
				{
				  "multi_match": {
					"type": "best_fields",
					"query": "%s",
					"lenient": true
				  }
				}
			  ],
			  "should": [],
			  "must_not": []
			}
		  }
		}`, query)

	req := api.SearchRequest{
		Index: []string{"anime"},
		Body:  bytes.NewReader([]byte(query)),
	}

	res, err := req.Do(context.TODO(), elasticClient)
	if err != nil {
		return nil, err
	}

	var animeRes []Anime
	var elasticRes ElasticResponse
	err = json.NewDecoder(res.Body).Decode(&elasticRes)
	if err != nil {
		return nil, err
	}
	fmt.Println(elasticRes)
	for i := range elasticRes.Hits.Hits {
		fmt.Println(elasticRes.Hits.Hits[i].ID)
		animeRes = append(animeRes, elasticRes.Hits.Hits[i].Source)
	}

	return animeRes, nil
}
