package twse

import (
	"database/sql"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IStockRepo interface {
	BatchInsertDividend([][]string, string) error
}

type StockRepo struct {
	db *gorm.DB
}

type Dividend struct {
	Code   int             `gorm:"primaryKey;column:CODE"`
	Date   string          `gorm:"primaryKey;column:DATE"`
	Name   sql.NullString  `gorm:"column:NAME"`
	PE     sql.NullFloat64 `gorm:"column:PE"`
	PB     sql.NullFloat64 `gorm:"column:PB"`
	Yield  sql.NullFloat64 `gorm:"column:YIELD"`
	Year   sql.NullString  `gorm:"column:YEAR"`
	Fiscal sql.NullString  `gorm:"column:FISCAL"`
}

func NewStockRepo(db *gorm.DB) IStockRepo {
	return &StockRepo{
		db: db,
	}
}

func (s *StockRepo) BatchInsertDividend(data [][]string, date string) error {
	var entiries []Dividend
	for _, v := range data {
		if len(v) == 8 {
			code, err := strconv.Atoi(v[0])
			if err != nil {
				return err
			}
			yield, err := s.toNullFloat64(v[2], "-", ",")
			if err != nil {
				return err
			}
			pe, err := s.toNullFloat64(v[4], "-", ",")
			if err != nil {
				return err
			}
			pb, err := s.toNullFloat64(v[5], "-", ",")
			if err != nil {
				return err
			}
			entiry := Dividend{
				Code: code,
				Date: date,
				Name: sql.NullString{
					Valid:  true,
					String: v[1],
				},
				Yield: yield,
				Year: sql.NullString{
					Valid:  true,
					String: v[3],
				},
				PE: pe,
				PB: pb,
				Fiscal: sql.NullString{
					Valid:  true,
					String: v[6],
				},
			}
			entiries = append(entiries, entiry)
		}
	}
	err := s.db.Clauses(clause.Insert{Modifier: "IGNORE"}).Table("DIVIDEND").Create(&entiries).Error
	return err
}

func (s *StockRepo) toNullFloat64(str, syntax, escape string) (sql.NullFloat64, error) {
	result := sql.NullFloat64{
		Valid: false,
	}
	if str == "" || str == syntax {
		return result, nil
	}
	if escape != "" {
		str = strings.ReplaceAll(str, escape, "")
	}
	f, err := strconv.ParseFloat(str, 10)
	if err != nil {
		return result, err
	}
	result.Float64 = f
	result.Valid = true
	return result, nil
}
