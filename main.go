package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func upload(c echo.Context) error {
	// read the passowrd.txt file
	pass, err := os.Open("password.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer pass.Close()
	// read the file
	b, err := ioutil.ReadAll(pass)
	if err != nil {
		log.Fatal(err)
	}
	password := c.FormValue("password")
	if password != string(b) {
		return echo.ErrUnauthorized
	}

	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	if file.Size > 100485760 || file.Header.Get("Content-Type") != "image/jpeg" && file.Header.Get("Content-Type") != "image/png" && file.Header.Get("Content-Type") != "video/mp4" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid file type or size (max 100MB)")
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

	// Copy
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
		json.NewEncoder(c.Response()).Encode(c.Scheme() + "://" + c.Request().Host + "/" + file.Name())

	}
	return nil
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "public")
	e.POST("/upload", upload)
	e.GET("/f", displayFiles)

	port, ok := os.LookupEnv("PORT")

	if ok == false {
		port = "5000"
	}
	fmt.Printf("server on port: %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
