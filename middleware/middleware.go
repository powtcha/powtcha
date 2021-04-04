package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/powtcha/powtcha"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	errPowtchaNotFound      = errors.New("powtcha: result not found")
	errPowtchaInvalidFormat = errors.New("powtcha: invalid result format")
)

type middleware struct {
	appID        uint32
	validity     time.Duration
	problems     byte
	difficulty   byte
	secret       []byte
	locationType string
	locationName string
	locationFunc func(c *gin.Context) (string, error)
}

func newPowtcha(config Config) *middleware {
	locationFunc := config.LocationFunc
	if locationFunc == nil {
		parts := strings.SplitN(config.Location, ":", 2)
		if len(parts) != 2 {
			panic("invalid location configuration: " + config.Location)
		}
		locationType := parts[0]
		switch locationType {
		case "header", "query", "form", "json":
			locationName := parts[1]
			locationFunc = func(c *gin.Context) (string, error) {
				switch locationType {
				case "header":
					return c.GetHeader(locationName), nil
				case "query":
					return c.Query(locationName), nil
				case "form":
					return c.PostForm(locationName), nil
				case "json":
					var body map[string]interface{}
					if err := c.BindJSON(body); err != nil {
						return "", err
					}
					location, ok := body[locationName]
					if !ok {
						return "", errPowtchaNotFound
					}
					locationStr, ok := location.(string)
					if !ok {
						return "", errPowtchaInvalidFormat
					}
					return locationStr, nil
				}
				return "", errPowtchaNotFound
			}
		default:
			panic("invalid location type: " + locationType)
		}
	}

	return &middleware{
		appID:        config.AppID,
		validity:       config.Validity,
		problems:     config.Problems,
		difficulty:   config.Difficulty,
		secret:       config.Secret,
		locationType: "",
		locationName: "",
		locationFunc: locationFunc,
	}
}

func (mw *middleware) Verify(c *gin.Context) {
	var err error
	var resultStr string
	if resultStr, err = mw.locationFunc(c); err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	log.Println(resultStr)
	result, err := powtcha.DecodeResult(resultStr, mw.secret)
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if !result.Valid(mw.appID) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.Status(http.StatusNoContent)
}

func (mw *middleware) Generate(c *gin.Context) {
	puzzle, err := powtcha.NewPuzzle(mw.appID, mw.validity, mw.problems, mw.difficulty)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	puzzleStr, err := puzzle.Encode(mw.secret)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.String(http.StatusOK, puzzleStr)
}

func New(config Config) *middleware {
	return newPowtcha(config)
}
