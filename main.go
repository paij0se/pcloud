package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	env "github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func goDotEnvVariable(key string) string {

	// load .env file
	err := env.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
func upload(c echo.Context) error {
	password := c.FormValue("password")
	// My Auth XD
	if password != goDotEnvVariable("PASSWORD") {
		return echo.ErrUnauthorized
	}
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	max, _ := strconv.ParseInt(os.Getenv("MAX_SIZE_FILE"), 10, 64) // Max 900mb file size
	if file.Size > max || file.Header.Get("Content-Type") != "image/jpeg" && file.Header.Get("Content-Type") != "image/png" && file.Header.Get("Content-Type") != "video/mp4" && file.Header.Get("Content-Type") != "image/gif" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid file type or size")
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create("public/" + file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}
	s := strings.ReplaceAll(c.Scheme()+"://"+c.Request().Host+""+dst.Name(), "public", "")
	json.NewEncoder(c.Response()).Encode(s)
	return nil

}
func displayFiles(c echo.Context) error {
	files, err := ioutil.ReadDir("public")

	if err != nil {
		log.Println(err)

	}

	for _, file := range files {
		switch file.Name()[strings.LastIndex(file.Name(), ".")+1:] {
		case "jpg", "jpeg", "png", "gif":
			c.HTML(http.StatusOK, "<img src='"+c.Scheme()+"://"+c.Request().Host+"/"+file.Name()+"'>"+"</img>")
		case "mp4":
			c.HTML(http.StatusOK, "<video controls><source src='"+c.Scheme()+"://"+c.Request().Host+"/"+file.Name()+"' type='video/mp4'></video>")
		default:
			c.HTML(http.StatusOK, "<a href='"+c.Scheme()+"://"+c.Request().Host+"/"+file.Name()+"'>"+file.Name()+"</a>")
		}
	}
	return nil
}

func main() {
	e := echo.New()
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(11)))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: 10, Burst: 30, ExpiresIn: 3 * time.Minute},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusForbidden, nil)
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, nil)
		},
	}
	e.Static("/", "public")
	e.POST("/upload", upload, middleware.RateLimiterWithConfig(config))
	e.GET("/f", displayFiles, middleware.RateLimiterWithConfig(config))
	port, ok := os.LookupEnv("PORT")

	if !ok {
		port = "5000"
	}
	fmt.Printf("server on port: %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
