package validator

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Normalizable interface {
	Normalize()
}

type Defaultable interface {
	ApplyDefaults()
}

type CrossValidatable interface {
	Validate() (field string, message string)
}

// ErrorHandlerFunc is the function signature for handling binding errors.
// It must be set before calling Bind functions. Typically set during app initialization.
var ErrorHandlerFunc func(c *gin.Context, err error)

func abortWithBindingError(c *gin.Context, err error) {
	if ErrorHandlerFunc != nil {
		ErrorHandlerFunc(c, err)
		return
	}
	// Fallback if not configured
	_ = c.Error(err)
	c.Abort()
}

func BindJSON(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		abortWithBindingError(c, err)
		return false
	}
	return postProcess(c, req)
}

func BindQuery(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindQuery(req); err != nil {
		abortWithBindingError(c, err)
		return false
	}
	return postProcess(c, req)
}

func BindURI(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindUri(req); err != nil {
		abortWithBindingError(c, err)
		return false
	}
	return true
}

func postProcess(c *gin.Context, req interface{}) bool {
	if n, ok := req.(Normalizable); ok {
		n.Normalize()
	}

	if d, ok := req.(Defaultable); ok {
		d.ApplyDefaults()
	}

	if v, ok := req.(CrossValidatable); ok {
		if field, msg := v.Validate(); field != "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Dữ liệu đầu vào không hợp lệ",
				"code":    "VALIDATION_ERROR",
				"errors": []gin.H{
					{"field": field, "message": msg},
				},
			})
			return false
		}
	}

	return true
}
