package elasticsearch

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jessewkun/gocommon/logger"
	"github.com/stretchr/testify/assert"
)

const testIndex = "test-gocommon-index"

var testClient *Client

// TestMain manages setup and teardown for all tests in this package.
func TestMain(m *testing.M) {
	logger.Cfg.Path = "./test.log"
	_ = logger.Init()

	if err := setup(); err != nil {
		log.Printf("skipping tests: elasticsearch setup failed: %v", err)
		// We don't os.Exit(1) here to allow other packages' tests to run
	}

	code := m.Run()

	if testClient != nil {
		teardown()
	}

	os.Remove("./test.log")
	os.Exit(code)
}

// setup initializes the elasticsearch client for tests.
func setup() error {
	Cfgs = map[string]*Config{
		"default": {
			Addresses:     []string{"http://127.0.0.1:9200"},
			IsLog:         true,
			SlowThreshold: 200,
		},
	}
	if err := Init(); err != nil {
		return err
	}
	client, err := GetConn("default")
	if err != nil {
		return err
	}
	testClient = client
	return nil
}

// teardown cleans up resources after tests.
func teardown() {
	if testClient != nil {
		// Use the new signature for DeleteIndex
		_ = testClient.DeleteIndex(context.Background(), testIndex)
	}
	_ = Close()
}

func skipIfNoES(t *testing.T) {
	if testClient == nil {
		t.Skip("skipping test: elasticsearch client not initialized")
	}
}

func TestManager(t *testing.T) {
	skipIfNoES(t)
	assert.NotNil(t, defaultManager)

	// Test GetConn
	c, err := GetConn("default")
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, testClient, c)

	_, err = GetConn("nonexistent")
	assert.Error(t, err)

	// Test HealthCheck
	health := HealthCheck()
	assert.NotNil(t, health["default"])
	assert.Equal(t, "success", health["default"].Status)
}

func TestClient_IndexManagement(t *testing.T) {
	skipIfNoES(t)
	ctx := context.Background()

	// Clean up at the end
	defer func() {
		_ = testClient.DeleteIndex(ctx, testIndex)
	}()

	// 1. Create Index
	mapping := map[string]any{"mappings": map[string]any{"properties": map[string]any{"name": map[string]any{"type": "text"}}}}
	err := testClient.CreateIndex(ctx, testIndex, mapping)
	assert.NoError(t, err)

	// 2. Index Exists
	exists, err := testClient.IndexExists(ctx, testIndex)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 3. Get Mapping
	var mappingResult map[string]interface{}
	err = testClient.GetIndexMapping(ctx, testIndex, &mappingResult)
	assert.NoError(t, err)
	assert.NotNil(t, mappingResult[testIndex]) // Check if the index key exists in the result

	// 4. Delete Index
	err = testClient.DeleteIndex(ctx, testIndex)
	assert.NoError(t, err)

	// 5. Index Should Not Exist
	exists, err = testClient.IndexExists(ctx, testIndex)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestClient_DocumentFlow(t *testing.T) {
	skipIfNoES(t)
	ctx := context.Background()

	// Ensure index exists
	mapping := map[string]any{"mappings": map[string]any{"properties": map[string]any{"name": map[string]any{"type": "text"}}}}
	_ = testClient.CreateIndex(ctx, testIndex, mapping)

	// Clean up at the end
	defer func() {
		_ = testClient.DeleteIndex(ctx, testIndex)
	}()

	type doc struct {
		Name string `json:"name"`
	}

	// 1. Index Document and force refresh
	myDoc := doc{Name: "gocommon"}
	err := testClient.Index(ctx, testIndex, "1", myDoc, "wait_for")
	assert.NoError(t, err)

	// 2. Get Document
	var docWrapper struct {
		Source doc `json:"_source"`
	}
	found, err := testClient.Get(ctx, testIndex, "1", &docWrapper)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "gocommon", docWrapper.Source.Name)

	// 3. Search Document
	query := map[string]any{"query": map[string]any{"match": map[string]any{"name": "gocommon"}}}
	var searchResult struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
		} `json:"hits"`
	}
	err = testClient.Search(ctx, testIndex, query, &searchResult)
	assert.NoError(t, err)
	assert.Equal(t, 1, searchResult.Hits.Total.Value)

	// 4. Delete Document
	err = testClient.Delete(ctx, testIndex, "1", "wait_for")
	assert.NoError(t, err)

	// 5. Get Deleted Document (should not be found)
	var deletedDocWrapper struct {
		Source doc `json:"_source"`
	}
	found, err = testClient.Get(ctx, testIndex, "1", &deletedDocWrapper)
	assert.NoError(t, err)
	assert.False(t, found)
}
