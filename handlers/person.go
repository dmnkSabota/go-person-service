package handlers

import (
	"errors"
	"net/http"
	"person-service/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PersonHandler struct {
	db *gorm.DB
}

func NewPersonHandler(db *gorm.DB) *PersonHandler {
	return &PersonHandler{db: db}
}

func (h *PersonHandler) SavePerson(c *gin.Context) {
	var req models.SavePersonRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
		return
	}

	var existingPerson models.Person
	if err := h.db.Where("external_id = ?", req.ExternalID).First(&existingPerson).Error; err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: "Person with this external_id already exists",
		})
		return
	}

	person := models.FromSaveRequest(req)

	if err := h.db.Create(&person).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to save person",
		})
		return
	}

	c.JSON(http.StatusCreated, person.ToResponse())
}

func (h *PersonHandler) GetPerson(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid ID format",
		})
		return
	}

	var person models.Person
	if err := h.db.First(&person, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Person not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve person",
		})
		return
	}

	c.JSON(http.StatusOK, person.ToResponse())
}
