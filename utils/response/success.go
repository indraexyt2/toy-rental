package response

import "github.com/gin-gonic/gin"

type APISuccessResponse struct {
	Status   int         `json:"status_code,omitempty"`
	Message  string      `json:"message,omitempty"`
	Metadata interface{} `json:"metadata,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

func ResponseSuccess(c *gin.Context, statusCode int, data, metadata interface{}, message string) {
	var newResponse = &APISuccessResponse{
		Status:   1,
		Data:     data,
		Metadata: metadata,
		Message:  message,
	}

	c.JSON(statusCode, newResponse)
}
