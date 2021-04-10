package twse

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
	"gorm.io/gorm"
)

type ITWSE interface {
	BWIBBU(string) error
}

type TWSE struct {
	url  string
	db   *gorm.DB
	repo IStockRepo
}

// NewTWSE instance
func NewTWSE(url string, db *gorm.DB) ITWSE {
	repo := NewStockRepo(db)
	return &TWSE{
		url:  url,
		db:   db,
		repo: repo,
	}
}

// BWIBBU 匯入個股日本益比 殖利率及股價淨值比
func (t *TWSE) BWIBBU(date string) error {
	url := fmt.Sprintf("%s/exchangeReport/BWIBBU_d?response=csv&date=%s&selectType=ALL", t.url, date)
	path := fmt.Sprintf("BWIBBU_d_ALL_%s.csv", date)
	err := t.downloadFile(url, path)
	if err != nil {
		return err
	}
	defer t.RemoveFile(path)
	data, err := t.readCSV(path, 2, 2)
	if err != nil {
		return err
	}
	if len(data) > 0 {
		return t.repo.BatchInsertDividend(data, date)
	}
	return nil
}

func (t *TWSE) downloadFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	return nil
}

func (t *TWSE) RemoveFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return err
}

func (t *TWSE) readCSV(path string, trimHead, trimFoot int) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	big5 := transform.NewReader(file, traditionalchinese.Big5.NewDecoder())
	r := csv.NewReader(big5)
	r.FieldsPerRecord = -1
	lines, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	e := len(lines)
	if e == 0 {
		return lines, nil
	}
	s := 0
	if trimHead > 0 && e > trimHead {
		s = trimHead
	}
	if trimFoot > 0 && e > trimFoot {
		e = e - trimFoot
	}
	return lines[s:e], err
}
