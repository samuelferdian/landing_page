package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/spf13/viper"
)

type MyForm struct {
	Msisdn       string
	Price        string
	Product      string
	Trxid        string
	Status       string
	ErrorCode    string
	ErrorMessage string
	ShowResult   bool
}

var mArr map[string]MyForm

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	t := &Template{
		templates: template.Must(template.ParseGlob("*.html")),
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Renderer = t

	myform := MyForm{}
	mArr = make(map[string]MyForm)

	e.GET("/", myform.Home)
	e.POST("/", myform.Input)
	e.GET("/result/", myform.ResultData)
	e.GET("/check/", myform.CheckValue)
	e.POST("/test/", myform.Test)

	e.Logger.Fatal(e.Start(":8888"))
}

func (m *MyForm) Home(c echo.Context) error {
	return c.Render(http.StatusOK, "form.html", m)
}

func (m *MyForm) Input(c echo.Context) error {
	m.Msisdn = c.FormValue("msisdn")
	m.Price = c.FormValue("price")
	m.Product = c.FormValue("product")
	m.Trxid = RandStringBytes(10)
	m.ShowResult = true

	_, ok := mArr[m.Msisdn]
	if !ok {
		mArr[m.Msisdn] = MyForm{
			m.Msisdn,
			m.Price,
			m.Product,
			m.Trxid,
			"PENDING",
			"",
			"",
			true,
		}
		m.Status = "PENDING"
	} else {
		m.Status = mArr[m.Msisdn].Status
	}

	viper.AddConfigPath(".")
	viper.SetConfigName("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	url := viper.Get("url")

	message := map[string]interface{}{
		"trx_id":  m.Trxid,
		"price":   m.Price,
		"msisdn":  m.Msisdn,
		"prod_id": m.Product,
		"op_code": "106",
	}

	bytesRepresentation, err := json.Marshal(message)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = http.Post(fmt.Sprintf("%v", url), "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(mArr)

	return c.Render(http.StatusOK, "form.html", m)
}

func (m *MyForm) ResultData(c echo.Context) error {
	// http://<ip>:<port>/<path>?trx_id=<trx_id>&result=<SUCCESS|FAILED>&err_code=<err_code>&err_msg=<err_msg>
	m.Trxid = c.FormValue("trx_id")
	m.Status = c.FormValue("result")
	m.ErrorCode = c.FormValue("err_code")
	m.ErrorMessage = c.FormValue("err_msg")

	fmt.Println(mArr)
	for _, v := range mArr {
		if v.Trxid == m.Trxid {
			fmt.Println(v.Trxid)
			mArr[v.Msisdn] = MyForm{
				v.Msisdn,
				v.Price,
				v.Product,
				m.Trxid,
				m.Status,
				m.ErrorCode,
				m.ErrorMessage,
				true,
			}
		}
	}
	fmt.Println(mArr)
	return c.JSON(http.StatusOK, mArr)
}

func (m *MyForm) CheckValue(c echo.Context) error {
	var res string
	for k, v := range mArr {
		if k == c.FormValue("msisdn") {
			res = v.Status
		}
	}
	return c.String(http.StatusOK, res)
}

func (m *MyForm) Test(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
