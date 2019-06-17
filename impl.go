package clickhouse

import (
	"context"
	"database/sql"
	"math"
	"reflect"
)

type mysqlImpl struct {
	*sql.DB
}

func newMysql(db *sql.DB) *mysqlImpl {
	return &mysqlImpl{
		DB: db,
	}
}

func (m *mysqlImpl) BeginTx(ctx context.Context) (ret Tx, err error) {
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	ret = &txImpl{
		Tx: tx,
	}
	return
}

func (m *mysqlImpl) Scan(psql string, srf ScanRowsFunc, args ...interface{}) (ret interface{}, err error) {

	pstmt, err := m.DB.Prepare(psql)
	if err != nil {
		return
	}
	rows, err := pstmt.Query(args...)
	if err != nil {
		return
	}
	defer rows.Close()

	return srf(rows)
}

func (m *mysqlImpl) ScanAll(psql string, srf ScanRowFunc, args ...interface{}) (ret interface{}, err error) {

	pstmt, err := m.DB.Prepare(psql)
	if err != nil {
		return
	}
	rows, err := pstmt.Query(args...)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		var (
			val   interface{}
			slice reflect.Value
		)
		val, err = srf(rows)
		if err != nil {
			return
		}
		slice = reflect.Append(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(val)), 0, InitialCapacity), reflect.ValueOf(val))

		for rows.Next() {
			val, err = srf(rows)
			if err != nil {
				return
			}
			slice = reflect.Append(slice, reflect.ValueOf(val))
		}
		ret = slice.Interface()
	}
	return
}

func (m *mysqlImpl) ScanOne2(psql string, to interface{}, args ...interface{}) (ok bool, err error) {

	pstmt, err := m.DB.Prepare(psql)
	if err != nil {
		return
	}
	rows, err := pstmt.Query(args...)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		switch to := to.(type) {
		case []interface{}:
			err = rows.Scan(to...)
		default:
			err = rows.Scan(to)
		}
		if err != nil {
			return
		}
		ok = true
	}
	return
}

func (m *mysqlImpl) ScanOne(psql string, srf ScanRowFunc, args ...interface{}) (ret interface{}, err error) {

	pstmt, err := m.DB.Prepare(psql)
	if err != nil {
		return
	}
	rows, err := pstmt.Query(args...)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		ret, err = srf(rows)
		if err != nil {
			return
		}
	}

	return
}

/*如果源SQL没有limit子句,则直接拼到最后即可*/
func (m *mysqlImpl) ScanRange(psql string, srf ScanRowFunc, offset int, limit int, args ...interface{}) (ret interface{}, err error) {
	meta := GetSqlMeta(psql)
	if meta.LimitPsql == "" {
		GenLimitSql(psql, meta)
	}
	args = append(args, offset, limit)

	pstmt, err := m.DB.Prepare(psql)
	if err != nil {
		return
	}
	rows, err := pstmt.Query(args...)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		var (
			val   interface{}
			slice reflect.Value
		)
		val, err = srf(rows)
		if err != nil {
			return
		}
		slice = reflect.Append(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(val)), 0, InitialCapacity), reflect.ValueOf(val))

		for rows.Next() {
			val, err = srf(rows)
			if err != nil {
				return
			}
			slice = reflect.Append(slice, reflect.ValueOf(val))
		}
		ret = slice.Interface()
	}
	return
}

func (m *mysqlImpl) ScanPage(psql string, srf ScanRowFunc, offset int, limit int, sort string, desc bool, args ...interface{}) (tot int, ret interface{}, err error) {

	aln := len(args)

	meta := GetSqlMeta(psql)

	// 查询记录
	dataPsql := GenDataSql(psql, meta, sort, desc)
	if limit <= 0 {
		limit = math.MaxInt32
	}
	args = append(args, offset, limit)

	pstmt, err := m.DB.Prepare(dataPsql)
	if err != nil {
		return
	}
	rows, err := pstmt.Query(args...)
	if err != nil {
		return
	}
	defer rows.Close()

	var dlen int
	if rows.Next() {
		var (
			val   interface{}
			slice reflect.Value
		)
		val, err = srf(rows)
		if err != nil {
			return
		}
		slice = reflect.Append(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(val)), 0, InitialCapacity), reflect.ValueOf(val))

		for rows.Next() {
			val, err = srf(rows)
			if err != nil {
				return
			}
			slice = reflect.Append(slice, reflect.ValueOf(val))
		}
		ret = slice.Interface()
		dlen = slice.Len()
	}
	if dlen == 0 && offset == 0 {
		tot = 0
	} else if dlen > 0 && dlen < limit {
		tot = offset + dlen
	} else {
		tot, err = m.scanPageTotal(psql, meta, args[0:aln]...)
	}

	return
}

func (m *mysqlImpl) scanPageTotal(psql string, meta *SqlMeta, args ...interface{}) (ret int, err error) {
	// 查询总数
	if meta.TotalPsql == "" {
		GenTotalSql(psql, meta)
	}

	pstmt, err := m.DB.Prepare(meta.TotalPsql)
	if err != nil {
		return
	}
	rows, err := pstmt.Query(args...)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&ret)
	}

	return
}

func (m *mysqlImpl) Exec(psql string, args ...interface{}) (ret sql.Result, err error) {
	pstmt, err := m.DB.Prepare(psql)
	if err != nil {
		return
	}
	defer pstmt.Close()
	ret, err = pstmt.Exec(args...)
	return
}

func (m *mysqlImpl) ExecBatch(psql string, argsList ...interface{}) (retList []sql.Result, err error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return
	}
	pstmt, err := tx.Prepare(psql)
	if err != nil {
		return
	}
	defer pstmt.Close()

	retList = make([]sql.Result, len(argsList))
	var ret sql.Result
	for i, args := range argsList {
		switch args := args.(type) {
		case []interface{}:
			ret, err = pstmt.Exec(args...)
		default:
			ret, err = pstmt.Exec(args)
		}
		if err != nil {
			tx.Rollback()
			return
		}
		retList[i] = ret
	}
	err = tx.Commit()
	return
}
