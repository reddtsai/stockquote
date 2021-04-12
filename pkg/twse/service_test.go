package twse

import (
	"encoding/json"
	"testing"

	"github.com/go-playground/assert"
)

type TestTWSE struct {
	twse *TWSE
}

type MockRepo int

func (repo *MockRepo) GetDividend(p1 string, p2 string, p3 ...interface{}) ([]Dividend, error) {
	data := `[
		{"code": 1101,"date": "20210409","name": "台泥","pe": "11.04","pb": "1.36","yield": "7.49","year": "109","fiscal": "109/4"},
        {"code": 1101,"date": "20210408","name": "台泥","pe": "11.09","pb": "1.36","yield": "7.46","year": "109","fiscal": "109/4"},
        {"code": 1101,"date": "20210407","name": "台泥","pe": "11.06","pb": "1.36","yield": "7.48","year": "109","fiscal": "109/4"}
    ]`
	dividend := []Dividend{}
	json.Unmarshal([]byte(data), &dividend)
	return dividend, nil
}
func (repo *MockRepo) GetDividendLimit(p1 string, p2 int, p3 string, p4 ...interface{}) ([]Dividend, error) {
	data := `[
		{"code": 1101,"date": "20210409","name": "台泥","pe": "11.04","pb": "1.36","yield": "7.49","year": "109","fiscal": "109/4"},
        {"code": 1101,"date": "20210408","name": "台泥","pe": "11.09","pb": "1.36","yield": "7.46","year": "109","fiscal": "109/4"},
        {"code": 1101,"date": "20210407","name": "台泥","pe": "11.06","pb": "1.36","yield": "7.48","year": "109","fiscal": "109/4"}
    ]`
	dividend := []Dividend{}
	json.Unmarshal([]byte(data), &dividend)
	return dividend, nil
}
func (repo *MockRepo) BatchInsertDividend(p1 [][]string, p2 string) error {
	return nil
}

func NewTestTWSE() *TestTWSE {
	repo := new(MockRepo)
	twse := &TWSE{
		url:  "",
		repo: repo,
	}
	t := &TestTWSE{
		twse: twse,
	}
	return t
}

func TestGetStock(t *testing.T) {
	instance := NewTestTWSE()
	record, err := instance.twse.GetStock(1101, 3)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(record), 3)
}

func TestRankingPE(t *testing.T) {
	instance := NewTestTWSE()
	record, err := instance.twse.RankingPE("20210409", 3)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, len(record), 3)
}

func TestGetStockUp(t *testing.T) {
	instance := NewTestTWSE()
	record, err := instance.twse.GetStockUp(1101, "20210406", "20210410")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, record.Code, 1101)
}
