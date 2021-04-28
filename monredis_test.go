package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/rwynn/gtm/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
This test requires the following processes to be running on localhost
	- elasticsearch v7.0+
	- mongodb 4.0+
	- monstache

monstache should be run with the following settings to force bulk requests
 -elasticsearch-max-docs 1
 -elasticsearch-max-seconds 1

WARNING: This test is destructive for the database test in mongodb and
any index prefixed with test in elasticsearch

If the tests are failing you can try increasing the delay between when
an operation in mongodb is checked in elasticsearch by passing the delay
argument (number of seconds; defaults to 5)

go test -v -delay 10
*/

var delay int

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var mongoURL = getEnv("MONGO_DB_URL", "mongodb://localhost:27017")

var elasticURL = getEnv("ELASTIC_SEARCH_URL", "http://localhost:9200")

func init() {
	testing.Init()
	fmt.Printf("MongoDB Url: %v\nElasticsearch Url: %v\n", mongoURL, elasticURL)

	flag.IntVar(&delay, "delay", 3, "Delay between operations in seconds")
	flag.Parse()
}

func dialMongo() (*mongo.Client, error) {
	rb := bson.NewRegistryBuilder()
	rb.RegisterTypeMapEntry(bsontype.DateTime, reflect.TypeOf(time.Time{}))
	reg := rb.Build()
	clientOptions := options.Client()
	clientOptions.ApplyURI(mongoURL)
	clientOptions.SetAppName("monstache")
	clientOptions.SetRegistry(reg)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}
	err = client.Connect(context.Background())
	if err != nil {
		return nil, err
	}
	return client, nil
}

func DropTestDB(t *testing.T, client *mongo.Client) {
	db := client.Database("test")
	if err := db.Drop(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestParseElasticsearchVersion(t *testing.T) {
	var err error
	c := &configOptions{}
	err = c.parseElasticsearchVersion("6.2.4")
	if err != nil {
		t.Fatal(err)
	}
	err = c.parseElasticsearchVersion("")
	if err == nil {
		t.Fatalf("Expected error for blank version")
	}
	err = c.parseElasticsearchVersion("0")
	if err == nil {
		t.Fatalf("Expected error for invalid version")
	}
}

func TestExtractRelateData(t *testing.T) {
	data, err := extractData("foo", map[string]interface{}{"foo": 1})
	if err != nil {
		t.Fatalf("Expected nil error")
	}
	if data != 1 {
		t.Fatalf("Expected extracting foo value of 1")
	}
	data, err = extractData("foo.bar", map[string]interface{}{"foo": map[string]interface{}{"bar": 1}})
	if err != nil {
		t.Fatalf("Expected nil error")
	}
	if data != 1 {
		t.Fatalf("Expected extracting foo.bar value of 1")
	}
	data, err = extractData("foo.bar", map[string]interface{}{"foo": map[string]interface{}{"foo": 1}})
	if err == nil {
		t.Fatalf("Expected error for missing key")
	}
	data, err = extractData("foo", map[string]interface{}{"bar": 1})
	if err == nil {
		t.Fatalf("Expected error for missing key")
	}
	data, err = extractData("foo.bar", map[string]interface{}{"foo": []string{"a", "b", "c"}})
	if err == nil {
		t.Fatalf("Expected error for missing key")
	}
	data, err = extractData("foo.bar", map[string]interface{}{"foo": map[string]interface{}{"bar": []map[string]interface{}{{"id": 1}, {"id": 2}, {"id": 3}, {"id": 4}}}})
	if err != nil {
		t.Fatalf("Expected nil error")
	}
	if data != [4]int{1, 2, 3, 4} {
		t.Fatalf("Expected extracting foo.bar value of 1")
	}
}

func TestBuildRelateSelector(t *testing.T) {
	sel := buildSelector("foo", 1)
	if sel == nil {
		t.Fatalf("Expected non-nil selector")
	}
	if len(sel) != 1 {
		t.Fatalf("Expected 1 foo key in selector")
	}
	if sel["foo"] != 1 {
		t.Fatalf("Expected matching foo to 1: %v", sel)
	}
	sel = buildSelector("foo.bar", 1)
	if sel == nil {
		t.Fatalf("Expected non-nil selector")
	}
	if len(sel) != 1 {
		t.Fatalf("Expected 1 foo key in selector")
	}
	bar, ok := sel["foo"].(bson.M)
	if !ok {
		t.Fatalf("Expected nested selector under foo")
	}
	if bar["bar"] != 1 {
		t.Fatalf("Expected matching foo.bar to 1: %v", sel)
	}
}

func TestOpIdToString(t *testing.T) {
	var result string
	var id = 10.0
	var id2 int64 = 1
	var id3 float32 = 12.0
	op := &gtm.Op{Id: id}
	result = opIDToString(op)
	if result != "10" {
		t.Fatalf("Expected decimal dropped from float64 for ID")
	}
	op.Id = id2
	result = opIDToString(op)
	if result != "1" {
		t.Fatalf("Expected int64 converted to string")
	}
	op.Id = id3
	result = opIDToString(op)
	if result != "12" {
		t.Fatalf("Expected int64 converted to string")
	}
}

func TestPruneInvalidJSON(t *testing.T) {
	ts := time.Date(-1, time.November, 10, 23, 0, 0, 0, time.UTC)
	m := make(map[string]interface{})
	m["a"] = math.Inf(1)
	m["b"] = math.Inf(-1)
	m["c"] = math.NaN()
	m["d"] = 1
	m["e"] = ts
	m["f"] = []interface{}{m["a"], m["b"], m["c"], m["d"], m["e"]}
	out := fixPruneInvalidJSON("docId-1", m)
	if len(out) != 2 {
		t.Fatalf("Expected 4 fields to be pruned")
	}
	if out["d"] != 1 {
		t.Fatalf("Expected 1 field to remain intact")
	}
	if len(out["f"].([]interface{})) != 1 {
		t.Fatalf("Expected 4 array fields to be pruned")
	}
	if out["f"].([]interface{})[0] != 1 {
		t.Fatalf("Expected 1 array field to remain intact")
	}
}
