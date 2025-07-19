package endpoint

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lavatee/astraltest/internal/model"
	"github.com/sirupsen/logrus"
)

func (e *Endpoint) UploadDocument(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	isFileLoaded := true
	if err != nil {
		isFileLoaded = false
	}
	defer file.Close()
	meta := c.PostForm("meta")
	jsonData := c.PostForm("json")
	token := c.GetString("token")
	doc, err := e.services.Documents.Upload(c.Request.Context(), token, meta, jsonData, file, header, isFileLoaded)
	if err != nil {
		logrus.Errorf("Failed to upload document (internal error): %s", err.Error())
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorInfo{Code: 500, Text: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, model.DataResponse{
		Data: doc,
	})
}

func (e *Endpoint) GetDocuments(c *gin.Context) {
	token := c.GetString("token")
	login := c.Query("login")
	key := c.Query("key")
	value := c.Query("value")
	limit, _ := strconv.Atoi(c.Query("limit"))
	docs, err := e.services.Documents.GetAll(c.Request.Context(), token, login, key, value, limit)
	if err != nil {
		logrus.Errorf("Failed to get documents (internal error): %s", err.Error())
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorInfo{Code: 500, Text: err.Error()},
		})
		return
	}
	if c.Request.Method == http.MethodHead {
		c.Status(http.StatusOK)
		return
	}
	c.JSON(http.StatusOK, model.DataResponse{
		Data: gin.H{"docs": docs},
	})
}

func (e *Endpoint) GetDocument(c *gin.Context) {
	id := c.Param("id")
	token := c.GetString("token")
	doc, fileData, err := e.services.Documents.GetByID(c.Request.Context(), token, id)
	if err != nil {
		logrus.Errorf("Failed to get document (internal error): %s", err.Error())
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorInfo{Code: 500, Text: err.Error()},
		})
		return
	}
	if c.Request.Method == http.MethodHead {
		c.Status(http.StatusOK)
		return
	}
	if doc.File {
		c.Data(http.StatusOK, doc.Mime, fileData)
	} else {
		c.JSON(http.StatusOK, model.DataResponse{
			Data: string(fileData),
		})
	}
}

func (e *Endpoint) DeleteDocument(c *gin.Context) {
	id := c.Param("id")
	token := c.GetString("token")
	err := e.services.Documents.Delete(c.Request.Context(), token, id)
	if err != nil {
		logrus.Errorf("Failed to delete document (internal error): %s", err.Error())
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorInfo{Code: 500, Text: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, model.Response{
		Response: gin.H{id: true},
	})
}
