package integration_v3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestSuite struct {
	ctx               context.Context
	postgresContainer testcontainers.Container
	menuServiceURL    string
	pgPool            *pgxpool.Pool
	menuServiceCmd    *exec.Cmd
}

type Category struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type MenuItem struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	IsAvailable   bool    `json:"is_available"`
	Quantity      int     `json:"quantity"`
	HasDiscount   bool    `json:"has_discount"`
	DiscountValue float64 `json:"discount_value"`
	CategoryID    string  `json:"category_id"`
}

func setupPostgres(ctx context.Context) (*postgres.PostgresContainer, string, error) {
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:14-alpine"),
		postgres.WithDatabase("menu"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, "", err
	}

	connStr, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		return nil, "", err
	}

	return pgContainer, connStr, nil
}

func setupMenuService(ctx context.Context, dbConnString string) (*exec.Cmd, string, error) {
	projectRoot, err := filepath.Abs("../../../")
	if err != nil {
		return nil, "", err
	}

	cmd := exec.CommandContext(ctx, "go", "run", filepath.Join(projectRoot, "cmd/main.go"))
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("DATABASE_URL=%s", dbConnString),
		"PORT=8083",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return nil, "", err
	}

	time.Sleep(10 * time.Second)

	return cmd, "http://localhost:8083", nil
}

func (ts *TestSuite) setupTestData() error {
	schemaSQL, err := os.ReadFile("../../../schema.sql")
	if err != nil {
		return err
	}

	_, err = ts.pgPool.Exec(ts.ctx, string(schemaSQL))
	return err
}

func setupSuite(t *testing.T) *TestSuite {
	ctx := context.Background()

	postgresContainer, pgConnString, err := setupPostgres(ctx)
	require.NoError(t, err)

	pgPool, err := pgxpool.New(ctx, pgConnString)
	require.NoError(t, err)

	menuServiceCmd, menuServiceURL, err := setupMenuService(ctx, pgConnString)
	require.NoError(t, err)

	ts := &TestSuite{
		ctx:               ctx,
		postgresContainer: postgresContainer,
		menuServiceURL:    menuServiceURL,
		pgPool:            pgPool,
		menuServiceCmd:    menuServiceCmd,
	}

	err = ts.setupTestData()
	require.NoError(t, err)

	return ts
}

func (ts *TestSuite) tearDown() {
	if ts.menuServiceCmd != nil && ts.menuServiceCmd.Process != nil {
		ts.menuServiceCmd.Process.Kill()
	}

	if ts.pgPool != nil {
		ts.pgPool.Close()
	}

	if ts.postgresContainer != nil {
		ts.postgresContainer.Terminate(ts.ctx)
	}
}

func TestMenuServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := setupSuite(t)
	defer ts.tearDown()

	t.Run("GetCategories", func(t *testing.T) {
		resp, err := http.Get(ts.menuServiceURL + "/api/categories")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string][]Category
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		categories := result["categories"]
		assert.NotEmpty(t, categories)
	})

	t.Run("GetMenu", func(t *testing.T) {
		resp, err := http.Get(ts.menuServiceURL + "/api/menu")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string][]map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		menu := result["menu"]
		assert.NotEmpty(t, menu)

		for _, category := range menu {
			assert.NotEmpty(t, category["name"])
			assert.NotEmpty(t, category["items"])
		}
	})
}

func TestMenuCRUDOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := setupSuite(t)
	defer ts.tearDown()

	var createdItemID string

	t.Run("CreateMenuItem", func(t *testing.T) {
		newItem := map[string]interface{}{
			"name":          "Integration Test Item",
			"price":         123.45,
			"category_name": "ام علي",
			"quantity":      25,
			"is_available":  true,
		}

		jsonData, err := json.Marshal(newItem)
		require.NoError(t, err)

		resp, err := http.Post(ts.menuServiceURL+"/api/menu", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		item, ok := result["item"].(map[string]interface{})
		assert.True(t, ok)

		createdItemID = item["id"].(string)
		assert.NotEmpty(t, createdItemID)
	})

	t.Run("GetMenuItem", func(t *testing.T) {
		resp, err := http.Get(ts.menuServiceURL + "/api/menu/" + createdItemID)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		item := result["item"]
		assert.Equal(t, createdItemID, item["id"])
		assert.Equal(t, "Integration Test Item", item["name"])
		assert.Equal(t, 123.45, item["price"])
	})

	t.Run("UpdateMenuItem", func(t *testing.T) {
		updateData := map[string]interface{}{
			"name":         "Updated Integration Test Item",
			"price":        99.99,
			"is_available": false,
		}

		jsonData, err := json.Marshal(updateData)
		require.NoError(t, err)

		client := &http.Client{}
		req, err := http.NewRequest("PUT", ts.menuServiceURL+"/api/menu/"+createdItemID, bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify the update
		getResp, err := http.Get(ts.menuServiceURL + "/api/menu/" + createdItemID)
		require.NoError(t, err)
		defer getResp.Body.Close()

		var result map[string]map[string]interface{}
		err = json.NewDecoder(getResp.Body).Decode(&result)
		require.NoError(t, err)

		item := result["item"]
		assert.Equal(t, "Updated Integration Test Item", item["name"])
		assert.Equal(t, 99.99, item["price"])
		assert.Equal(t, false, item["is_available"])
	})

	t.Run("AddDiscount", func(t *testing.T) {
		discountData := map[string]interface{}{
			"discount_value": 20.0,
		}

		jsonData, err := json.Marshal(discountData)
		require.NoError(t, err)

		client := &http.Client{}
		req, err := http.NewRequest("POST", ts.menuServiceURL+"/api/menu/"+createdItemID+"/discount", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify the discount was applied
		getResp, err := http.Get(ts.menuServiceURL + "/api/menu/" + createdItemID)
		require.NoError(t, err)
		defer getResp.Body.Close()

		var result map[string]map[string]interface{}
		err = json.NewDecoder(getResp.Body).Decode(&result)
		require.NoError(t, err)

		item := result["item"]
		assert.Equal(t, true, item["has_discount"])
		assert.Equal(t, 20.0, item["discount_value"])
		assert.InDelta(t, 79.992, item["effective_price"], 0.001) // 99.99 - (99.99 * 0.2)
	})

	t.Run("RemoveDiscount", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("DELETE", ts.menuServiceURL+"/api/menu/"+createdItemID+"/discount", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify the discount was removed
		getResp, err := http.Get(ts.menuServiceURL + "/api/menu/" + createdItemID)
		require.NoError(t, err)
		defer getResp.Body.Close()

		var result map[string]map[string]interface{}
		err = json.NewDecoder(getResp.Body).Decode(&result)
		require.NoError(t, err)

		item := result["item"]
		assert.Equal(t, false, item["has_discount"])
		assert.Equal(t, 0.0, item["discount_value"])
		assert.Equal(t, 0.0, item["effective_price"])
	})

	t.Run("DeleteMenuItem", func(t *testing.T) {
		client := &http.Client{}
		req, err := http.NewRequest("DELETE", ts.menuServiceURL+"/api/menu/"+createdItemID, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify the item was deleted
		getResp, err := http.Get(ts.menuServiceURL + "/api/menu/" + createdItemID)
		require.NoError(t, err)
		defer getResp.Body.Close()

		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})
}
