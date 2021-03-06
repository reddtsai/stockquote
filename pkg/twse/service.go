package twse

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
	"gorm.io/gorm"
)

type ITWSE interface {
	ImportStock(string) error
	GetStock(int, int) ([]Stock, error)
	GetStockUp(int, string, string) (StockUp, error)
	RankingPE(string, int) ([]PE, error)
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

type Stock struct {
	Code   int    `json:"code"`
	Date   string `json:"date"`
	Name   string `json:"name"`
	PE     string `json:"pe"`
	PB     string `json:"pb"`
	Yield  string `json:"yield"`
	Year   string `json:"year"`
	Fiscal string `json:"fiscal"`
}

func (t *TWSE) GetStock(code, count int) ([]Stock, error) {
	stocks := []Stock{}
	record, err := t.repo.GetDividendLimit("DATE desc", count, "CODE = ?", code)
	if err != nil {
		return stocks, err
	}
	for _, v := range record {
		stock := Stock{
			Code:   v.Code,
			Date:   v.Date,
			Name:   "-",
			PE:     "-",
			PB:     "-",
			Yield:  "-",
			Year:   "-",
			Fiscal: "-",
		}
		if v.Name.Valid {
			stock.Name = v.Name.String
		}
		if v.PE.Valid {
			stock.PE = strconv.FormatFloat(v.PE.Float64, 'f', -1, 64)
		}
		if v.PB.Valid {
			stock.PB = strconv.FormatFloat(v.PB.Float64, 'f', -1, 64)
		}
		if v.Yield.Valid {
			stock.Yield = strconv.FormatFloat(v.Yield.Float64, 'f', -1, 64)
		}
		if v.Year.Valid {
			stock.Year = v.Year.String
		}
		if v.Fiscal.Valid {
			stock.Fiscal = v.Fiscal.String
		}
		stocks = append(stocks, stock)
	}
	return stocks, nil
}

type PE struct {
	Code int    `json:"code"`
	Date string `json:"date"`
	Name string `json:"name"`
	PE   string `json:"pe"`
}

func (t *TWSE) RankingPE(date string, count int) ([]PE, error) {
	rank := []PE{}
	record, err := t.repo.GetDividendLimit("PE desc", count, "DATE = ?", date)
	if err != nil {
		return rank, err
	}
	for _, v := range record {
		pe := PE{
			Code: v.Code,
			Date: v.Date,
			Name: "-",
			PE:   "-",
		}
		if v.Name.Valid {
			pe.Name = v.Name.String
		}
		if v.PE.Valid {
			pe.PE = strconv.FormatFloat(v.PE.Float64, 'f', -1, 64)
		}
		rank = append(rank, pe)
	}
	return rank, nil
}

type F64 struct {
	Val float64
}

type StockUp struct {
	Code  int    `json:"code"`
	Days  int    `json:"days"`
	Begin string `json:"begin"`
	End   string `json:"end"`
}

func (t *TWSE) GetStockUp(code int, begin, end string) (StockUp, error) {
	stock := StockUp{
		Code: code,
	}
	record, err := t.repo.GetDividend("DATE", "CODE = ? AND DATE BETWEEN ? AND ?", code, begin, end)
	if err != nil {
		return stock, err
	}
	var tmp *F64
	tmpUP := StockUp{}
	for k, v := range record {
		if k == 0 {
			tmpUP.Begin = v.Date
			tmpUP.End = v.Date
			tmpUP.Days = 1
			stock.Begin = tmpUP.Begin
			stock.End = tmpUP.End
			stock.Days = tmpUP.Days
		} else {
			if tmp != nil && v.Yield.Valid {
				d := (v.Yield.Float64 - tmp.Val)
				if d > 0 {
					tmpUP.Days++
					tmpUP.End = v.Date
				} else {
					if tmpUP.Days > stock.Days {
						stock.Begin = tmpUP.Begin
						stock.End = tmpUP.End
						stock.Days = tmpUP.Days
					}
					tmpUP.Begin = v.Date
					tmpUP.End = v.Date
					tmpUP.Days = 1
				}
			}
		}
		tmp = nil
		if v.Yield.Valid {
			tmp = &F64{
				Val: v.Yield.Float64,
			}
		}
	}
	return stock, nil
}

func (t *TWSE) ImportStock(date string) error {
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
