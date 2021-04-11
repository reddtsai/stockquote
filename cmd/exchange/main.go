package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/reddtsai/stockquote/pkg/logger"
	"github.com/reddtsai/stockquote/pkg/twse"
)

func main() {
	logger.Info("start!!")

	// db
	dbHost := flag.String("db-host", "127.0.0.1", "db host")
	flag.Parse()
	dsn := fmt.Sprintf("root:123456@tcp(%s:3306)/STOCK?charset=utf8mb4&parseTime=True&loc=Local", *dbHost)
	sql := mysql.New(mysql.Config{
		DSN: dsn,
	})
	db, err := gorm.Open(sql)
	if err != nil {
		logger.Fatal(err)
	}

	// twse
	twse := twse.NewTWSE("https://www.twse.com.tw", db)

	// http server
	g := newServer(twse)
	srv := &http.Server{
		Addr:    ":5000",
		Handler: g.router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctxt, timeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer timeout()
	if err := srv.Shutdown(ctxt); err != nil {
		logger.Error(err)
	}

	logger.Info("stop!!")
}

type Server struct {
	twse   twse.ITWSE
	router *gin.Engine
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func newServer(twse twse.ITWSE) *Server {
	srv := &Server{
		twse: twse,
	}
	router := gin.Default()
	router.Use(cors.Default())
	v1 := router.Group("/v1")
	{
		v1.GET("/stock/:code", srv.stock)
		v1.GET("/ranking/:date", srv.ranking)
		v1.GET("/dividend", srv.dividend)
		v1.GET("/download/:date", srv.download)
	}
	srv.router = router
	return srv
}

func (s *Server) stock(c *gin.Context) {
	resp := &Response{
		Code: 9,
		Data: []twse.Stock{},
	}
	codePath := c.Param("code")
	code, err := strconv.Atoi(codePath)
	if err != nil {
		resp.Message = "stock code is not valid!!"
		c.JSON(400, resp)
		return
	}
	cntQuerystring := c.DefaultQuery("count", "10")
	cnt, err := strconv.Atoi(cntQuerystring)
	if err != nil {
		resp.Message = "stock count is not valid!!"
		c.JSON(400, resp)
		return
	}
	record, err := s.twse.GetStock(code, cnt)
	if err != nil {
		resp.Message = "system error!!"
		c.JSON(500, resp)
		return
	}

	resp.Code = 0
	resp.Data = record
	c.JSON(200, resp)
}

func (s *Server) ranking(c *gin.Context) {
	resp := &Response{
		Code: 9,
		Data: []twse.PE{},
	}
	date := c.Param("date")
	cntQuerystring := c.DefaultQuery("count", "10")
	cnt, err := strconv.Atoi(cntQuerystring)
	if err != nil {
		resp.Message = "rank count is not valid!!"
		c.JSON(400, resp)
		return
	}
	record, err := s.twse.RankingPE(date, cnt)
	if err != nil {
		resp.Message = "system error!!"
		c.JSON(500, resp)
		return
	}

	resp.Code = 0
	resp.Data = record
	c.JSON(200, resp)
}

func (s *Server) dividend(c *gin.Context) {

}

func (s *Server) download(c *gin.Context) {
	resp := &Response{
		Code: 9,
		Data: []twse.PE{},
	}
	date := c.Param("date")
	client, err := rpc.DialHTTP("tcp", "app-mining:5050")
	if err != nil {
		logger.Error(err)
		resp.Message = "system error!!"
		c.JSON(500, resp)
		return
	}
	result := ""
	err = client.Call("Server.ImportStock", date, &result)
	if err != nil {
		resp.Message = "system error!!"
		c.JSON(500, resp)
		return
	}

	resp.Code = 0
	resp.Message = date
	c.JSON(200, resp)
}
