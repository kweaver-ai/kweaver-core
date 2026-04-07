package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Recovery(log *logrus.Entry) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.WithField("error", err).Error("Panic recovered")

				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500000,
					"message": "Internal server error",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
