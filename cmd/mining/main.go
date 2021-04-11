package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// twse
	s := twse.NewTWSE("https://www.twse.com.tw", db)

	// rpc
	NewRPC(s)

	// import task
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
Loop:
	for {
		select {
		case <-ticker.C:
			n := time.Now()
			date := fmt.Sprintf("%d%d%d", n.Year(), n.Month(), n.Day()-1)
			logger.Info("ImportStock", date)
			err = s.ImportStock(date)
			if err != nil {
				logger.Error(err)
			}
		case <-quit:
			break Loop
		}
	}

	logger.Info("stop!!")
}

type Server struct {
	twse twse.ITWSE
}

func NewRPC(twse twse.ITWSE) {
	srv := &Server{
		twse: twse,
	}
	rpc.Register(srv)
	rpc.HandleHTTP()

	go func() {
		if err := http.ListenAndServe(":5050", nil); err != nil && err != http.ErrServerClosed {
			logger.Error(err)
		}
	}()
}

func (s *Server) ImportStock(date string, reply *string) error {
	logger.Info("ImportStock", date)
	err := s.twse.ImportStock(date)
	if err != nil {
		logger.Error(err)
		return err
	}
	reply = &date
	return nil
}
