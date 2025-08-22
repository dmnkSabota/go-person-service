package tests

import (
	"bytes"
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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	router *gin.Engine
	db     *gorm.DB
)

func TestMain(m *testing.M) {
	if shouldSkipTests() {
		fmt.Println("Skipping tests - PostgreSQL not available")
		fmt.Println("To run tests, start PostgreSQL with: docker-compose up -d postgres")
		os.Exit(0)
	}

	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func shouldSkipTests() bool {
	dbURL := "postgres://user:password@localhost:5432/persons?sslmode=disable"
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		return true
	}

	sqlDB, _ := db.DB()
	sqlDB.Close()
	return false
}

func setup() {
	gin.SetMode(gin.TestMode)

	dbURL := "postgres://user:password@localhost:5432/persons?sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	if err := db.AutoMigrate(&models.Person{}); err != nil {
		panic("Failed to migrate test database: " + err.Error())
	}

	personHandler := handlers.NewPersonHandler(db)
	router = gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.POST("/save", personHandler.SavePerson)
	router.GET("/:id", personHandler.GetPerson)
}

func teardown() {
	if db != nil {
		cleanTestData()
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
