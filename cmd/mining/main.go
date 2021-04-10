package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	// dsn = "root:123456@tcp(db:3306)/STOCK?charset=utf8mb4&parseTime=True&loc=Local"
	sql := mysql.New(mysql.Config{
		DSN: dsn,
	})
	db, err := gorm.Open(sql)
	if err != nil {
		log.Fatal(err)
	}

	// import task
	// TODO
	// 	ticker := time.NewTicker(5 * time.Second)
	// 	defer ticker.Stop()
	// Loop:
	// 	for {
	// 		select {
	// 		case <-ticker.C:
	// 			hub.pushBalance()
	// 		case <-hub.ctx.Done():
	// 			break Loop
	// 		}
	// 	}

	s := twse.NewTWSE("https://www.twse.com.tw", db)
	err = s.BWIBBU("20210408")
	// err = s.BWIBBU(time.Now().Format("20060102"))
	if err != nil {
		log.Println(err)
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("stop!!")
}
