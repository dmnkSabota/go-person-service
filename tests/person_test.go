package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"person-service/handlers"
	"person-service/models"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgresContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	router    *gin.Engine
	db        *gorm.DB
	container *postgresContainer.PostgresContainer
	ctx       context.Context
)

func TestMain(m *testing.M) {
	ctx = context.Background()

	if !isDockerAvailable() {
		fmt.Println("Skipping tests - Docker not available")
		fmt.Println("Install Docker to run integration tests")
		os.Exit(0)
	}

	if err := setupTestContainer(); err != nil {
		fmt.Printf("Failed to setup test container: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	teardown()
	os.Exit(code)
}

func isDockerAvailable() bool {
	testCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:      "hello-world",
		WaitingFor: wait.ForExit(),
	}

	container, err := testcontainers.GenericContainer(testCtx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          false,
	})

	if err != nil {
		return false
	}

	defer func() {
		if container != nil {
			err := container.Terminate(testCtx)
			if err != nil {
				return
			}
		}
	}()

	return true
}

func setupTestContainer() error {
	gin.SetMode(gin.TestMode)

	var err error
	container, err = postgresContainer.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgresContainer.WithDatabase("persons_test"),
		postgresContainer.WithUsername("testuser"),
		postgresContainer.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to get connection string: %w", err)
	}

	db, err = gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to test database: %w", err)
	}

	if err := db.AutoMigrate(&models.Person{}); err != nil {
		return fmt.Errorf("failed to migrate test database: %w", err)
	}

	personHandler := handlers.NewPersonHandler(db)
	router = gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.POST("/save", personHandler.SavePerson)
	router.GET("/:id", personHandler.GetPerson)

	return nil
}

func teardown() {
	if container != nil {
		if err := container.Terminate(ctx); err != nil {
			fmt.Printf("Failed to terminate container: %v\n", err)
		}
	}
}

func cleanTestData() {
	if db != nil {
		db.Where("name LIKE ? OR name LIKE ?", "Test%", "%Test%").Delete(&models.Person{})
	}
}

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestSavePersonSuccess(t *testing.T) {
	cleanTestData()

	externalID := uuid.New()
	reqBody := models.SavePersonRequest{
		ExternalID:  externalID,
		Name:        "Test User John",
		Email:       "testjohn@example.com",
		DateOfBirth: time.Date(1990, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/save", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.PersonResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, externalID, response.ExternalID)
	assert.Equal(t, "Test User John", response.Name)
}

func TestSavePersonDuplicateExternalID(t *testing.T) {
	cleanTestData()

	externalID := uuid.New()

	person1 := models.Person{
		ExternalID:  externalID,
		Name:        "Test First Person",
		Email:       "testfirst@example.com",
		DateOfBirth: time.Now(),
	}
	err := db.Create(&person1).Error
	require.NoError(t, err)

	reqBody := models.SavePersonRequest{
		ExternalID:  externalID,
		Name:        "Test Second Person",
		Email:       "testsecond@example.com",
		DateOfBirth: time.Now(),
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/save", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Contains(t, errorResponse.Error, "already exists")
}

func TestGetPersonSuccess(t *testing.T) {
	cleanTestData()

	person := models.Person{
		ExternalID:  uuid.New(),
		Name:        "Test Jane Doe",
		Email:       "testjane@example.com",
		DateOfBirth: time.Date(1985, 6, 15, 10, 30, 0, 0, time.UTC),
	}
	err := db.Create(&person).Error
	require.NoError(t, err)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%d", person.ID), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.PersonResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, person.ExternalID, response.ExternalID)
	assert.Equal(t, "Test Jane Doe", response.Name)
}

func TestGetPersonNotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "/999999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Person not found", errorResponse.Error)
}

func TestGetPersonInvalidID(t *testing.T) {
	req := httptest.NewRequest("GET", "/invalid-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Invalid ID format", errorResponse.Error)
}

func TestSavePersonInvalidEmail(t *testing.T) {
	cleanTestData()

	externalID := uuid.New()
	reqBody := models.SavePersonRequest{
		ExternalID:  externalID,
		Name:        "Test User",
		Email:       "invalid-email",
		DateOfBirth: time.Date(1990, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/save", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Contains(t, errorResponse.Error, "Invalid request")
}

func TestSavePersonMissingFields(t *testing.T) {
	cleanTestData()

	reqBody := map[string]interface{}{
		"external_id":   uuid.New(),
		"email":         "test@example.com",
		"date_of_birth": "1990-01-01T12:00:00Z",
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/save", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Contains(t, errorResponse.Error, "Invalid request")
}
