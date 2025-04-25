package controller

import (
	"errors"
	"final-project/entity"
	"final-project/service"
	"final-project/utils/helpers"
	"final-project/utils/response"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
	"math"
	"net/http"
	"path/filepath"
)

type IToyImageController interface {
	FindAll(c *gin.Context)
	Insert(c *gin.Context)
	DeleteById(c *gin.Context)
}

type ToyImageController struct {
	toyImageSvc service.IToyImageService
}

func NewToyImageController(toyImageSvc service.IToyImageService) IToyImageController {
	return &ToyImageController{toyImageSvc: toyImageSvc}
}

// FindAll godoc
// @Summary Mengambil semua data toy image
// @Tags Toy Image
// @Produce json
// @Param page query string false "Page"
// @Param limit query string false "Limit"
// @Success 200 {object} entity.ToyImage
// @Router /toy/image [get]
func (t ToyImageController) FindAll(c *gin.Context) {
	var logger = helpers.Logger

	var page = c.DefaultQuery("page", "1")
	var pageInt = helpers.ParseToInt(page)

	var limit = c.DefaultQuery("limit", "10")
	var limitInt = helpers.ParseToInt(limit)

	var offset = (pageInt - 1) * limitInt

	data, totalData, err := t.toyImageSvc.FindAll(c.Request.Context(), limitInt, offset)
	if err != nil {
		logger.Error("Failed to find all toy images: ", err)
		response.ResponseError(c, http.StatusInternalServerError, "Failed to find all toy images")
		return
	}

	metaData := response.Page{
		Limit:     limitInt,
		Total:     int(totalData),
		Page:      pageInt,
		TotalPage: int(math.Ceil(float64(totalData) / float64(limitInt))),
	}

	response.ResponseSuccess(c, http.StatusOK, data, metaData, "Success to find all toy images")
}

// Insert godoc
// @Summary Insert toy image
// @Tags Toy Image
// @Accept json
// @Produce json
// @Param images formData file true "Upload multiple images"
// @Success 200 {object} entity.ToyImage
// @Router /toy/image [post]
func (t ToyImageController) Insert(c *gin.Context) {
	var logger = helpers.Logger

	form, err := c.MultipartForm()
	if err != nil {
		logger.Error("Failed to get multipart form: ", err)
		response.ResponseError(c, http.StatusBadRequest, "Failed to get multipart form")
		return
	}

	images := form.File["images"]
	if len(images) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setidaknya satu gambar diperlukan"})
		return
	}

	for _, image := range images {
		id, _ := uuid.NewV7()
		extension := filepath.Ext(image.Filename)
		filename := fmt.Sprintf("%s%s", id, extension)
		savePath := filepath.Join("uploads", filename)

		if err := c.SaveUploadedFile(image, savePath); err != nil {
			logger.Error("Failed to save image: ", err)
			continue
		}

		err = t.toyImageSvc.Insert(c.Request.Context(), &entity.ToyImage{
			ImageURL: savePath,
		})
		if err != nil {
			logger.Error("Failed to insert toy image: ", err)
			response.ResponseError(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	response.ResponseSuccess(c, http.StatusOK, nil, nil, "Success to insert toy image")
}

// DeleteById godoc
// @Summary Delete toy berdasarkan id
// @Tags Toy Image
// @Produce json
// @Param id path string true "Toy Image ID"
// @Success 200 {object} entity.ToyImage
// @Router /toy/image/{id} [delete]
func (t ToyImageController) DeleteById(c *gin.Context) {
	var logger = helpers.Logger

	var id = c.Param("id")
	if id == "" {
		logger.Error("Id is required")
		response.ResponseError(c, http.StatusBadRequest, "Id is required")
		return
	}

	err := t.toyImageSvc.DeleteById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error(fmt.Errorf("toy image with id %s not found", id))
			response.ResponseError(c, http.StatusNotFound, "Toy image not found")
			return
		}

		logger.Error(fmt.Errorf("failed to delete toy image by id %s: %v", id, err))
		response.ResponseError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, nil, nil, "Success delete toy image")
}
