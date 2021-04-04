package middleware

import (
	"github.com/gin-gonic/gin"
	"time"
)

type Config struct {
	AppID      uint32
	Validity   time.Duration
	Problems   byte
	Difficulty byte
	Secret     []byte
	/**
	Location of the encoded result as "$type:$name" with the following types
	header - Contained in a header named $name
	query - Contained in a GET parameter named $name
	form - Contained in a POST form variable named $name
	json - Contained in a JSON variable named $name (currently only top level supported)
	*/
	Location     string // "body:$name"
	LocationFunc func(c *gin.Context) (string, error)
}
