# package clickhouse
clickhouse客户端

# Installation
- go get
```
go get -u github.com/kshvakov/clickhouse
go get -u github.com/obase/clickhouse
```
- go mod
```
go mod edit -require=github.com/obase/clickhouse@latest
```
自动级联下载相关依赖

# Configuration
conf.yml
```
clickhouse:
  -
    # 引用的key(必需)
    key: demo
    # 地址(必需). 多值用逗号分隔
    address: "10.11.165.44:9000"
    # DB名字(必需)
    database: "demo"
    # 用户名(可选)
    username: ""
    # 密码(可选)
    password: ""
    # 最大空闲数量(可选)
    maxIdleConns:
    # 最大打开数量(可选)
    maxOpenConns:
    # 最大lifetime(可选)
    connMaxLifetime: "0s"
    # 读超时(秒)
    read_timeout: 10
    # 写超时(秒)
    write_timeout : 10
    # 无延迟
    no_delay: true
    # connection_open_strategy 连接打开策略
    connection_open_strategy: "random"
    # 读写块大小
    block_size: 1000000
    # 连接池大小
    pool_size: 100
    # 是否debug
    debug: false
```

# Index
- Constants
```
const InitialCapacity = 256
```
- Variables
- type ScanRowFunc
```
/*
用于Rows.Scan()使用, 并返回解析结果. 必须注意:
- 参数cache用于对应当次rows的可重用缓存,避免反复创建导致GC! 如果cacheo==nil表明无需cache!. 该参数一般情况不需用到!
- 结果ret不能是nil, 否则反射报错!
*/
type ScanRowFunc func(row *sql.Rows) (interface{}, error)
```
- type ScanRowsFunc
```
/*
也用于Rows.Scan()使用, 并返回全部解析结果. 由用户自定义解析过程, 所以没有ScanRowFunc的局限!
*/
type ScanRowsFunc func(rows *sql.Rows) (interface{}, error)
```
- type Operation
```
type Operation interface {
	// 用户自定义解析过程
	Scan(psql string, srf ScanRowsFunc, args ...interface{}) (ret interface{}, err error)
	// 根据第一条数据反射结果, 要求首条数据结果不能为nil.
	ScanAll(psql string, srf ScanRowFunc, args ...interface{}) (ret interface{}, err error)
	ScanOne2(psql string, ret interface{}, args ...interface{}) (ok bool, err error)
	ScanOne(psql string, srf ScanRowFunc, args ...interface{}) (ret interface{}, err error)
	ScanRange(psql string, srf ScanRowFunc, offset int, limit int, args ...interface{}) (ret interface{}, err error)
	ScanPage(psql string, srf ScanRowFunc, offset int, limit int, sort string, desc bool, args ...interface{}) (tot int, ret interface{}, err error)
	scanPageTotal(psql string, meta *SqlMeta, args ...interface{}) (ret int, err error)

	Exec(psql string, args ...interface{}) (ret sql.Result, err error)
	ExecBatch(psql string, argsList ...interface{}) (retList []sql.Result, err error)
}
```
- func (m *clickhImpl) Scan
```
func (m *clickhImpl) Scan(psql string, srf ScanRowsFunc, args ...interface{}) (ret interface{}, err error) 
```
查询psql并将结果应用到srf函数. 各参数意义:
```
- psql: 查询SQL, 参数用?表示
- srf: SanRowsFunc, 封装查询结果的处理并返回最终结果
- args: 对应psql里面?的实参
```
注意: Scan函数返回的结果就是ScanRowsFunc的结果

- func (m *clickhImpl) ScanAll
```
func (m *clickhImpl) ScanAll(psql string, srf ScanRowFunc, args ...interface{}) (ret interface{}, err error) 
```
查询psql并将每一行记录应用到srf函数. 各参数意义:
```
- psql: 查询SQL, 参数用?表示
- srf: SanRowFunc, 封装查询结果的处理并返回最终结果
- args: 对应psql里面?的实参
```
注意： ScanAll函数返回的结果是[]T, 其中T是处理第一行记录时ScanRowFunc的返回结果类型. 如果没有任何记录, 即srf无法调用, 最后结果是nil

- func (m *clickhImpl) ScanOne
```
func (m *clickhImpl) ScanOne(psql string, srf ScanRowFunc, args ...interface{}) (ret interface{}, err error) 
```
查询psql的第一行记录并应用到srf函数. 各参数意义:
```
- psql: 查询SQL, 参数用?表示
- srf: SanRowFunc, 封装查询结果的处理并返回最终结果
- args: 对应psql里面?的实参
```
注意: ScanOne函数只处理第一条记录且返回的结果是T, 其中T是处理第一行记录时ScanRowFunc的返回结果类型. 如果没有任何记录, 即srf无法调用, 最后结果是nil

- func (m *clickhImpl) ScanOne2
```
func (m *clickhImpl) ScanOne2(psql string, to interface{}, args ...interface{}) (ok bool, err error)
```
查询psql的第一行记录并应用到srf函数. 各参数意义:
各参数意义:
```
- psql: 查询SQL, 参数用?表示
- to: 目标结果变量. 必须是指针类型
- srf: SanRowFunc, 封装查询结果的处理并返回最终结果
- args: 对应psql里面?的实参
```
注意: ScanOne2函数只处理第一条记录且返回的结果是T, 其中T是处理第一行记录时ScanRowFunc的返回结果类型. 如果没有任何记录, 即srf无法调用, 最后结果是nil

- func (m *clickhImpl) ScanRange
```
func (m *clickhImpl) ScanRange(psql string, srf ScanRowFunc, offset int, limit int, args ...interface{}) (ret interface{}, err error)
```
查询psql的某个范围并应用到srf函数. 各参数意义:
```
- psql: 查询SQL, 参数用?表示
- srf: SanRowFunc, 封装查询结果的处理并返回最终结果
- offset: 对应SQL的limit offset, count 
- limit: 对应SQL的limit offset, count 
- args: 对应psql里面?的实参
```
注意: ScanRange函数处理某个范围的记录且返回的结果是T, 其中T是处理第一行记录时ScanRowFunc的返回结果类型. 如果没有任何记录, 即srf无法调用, 最后结果是nil

- func (m *clickhImpl) ScanPage
```
func (m *clickhImpl) ScanPage(psql string, srf ScanRowFunc, offset int, limit int, sort string, desc bool, args ...interface{}) (tot int, ret interface{}, err error) 
```
通用分页查询函数.各参数意义:
```
- psql: 查询SQL, 参数用?表示
- srf: SanRowFunc, 封装查询结果的处理并返回最终结果
- offset: 对应SQL的limit offset, count 
- limit: 对应SQL的limit offset, count 
- sort: 排序字段,必须是select子句指定的字段
- desc: 是否DESC排序, false表示ASC, true表示DESC
- args: 对应psql里面?的实参
- tot: 对应SQL的总记录数
- ret: 对应SQL的[]T,即分页查询结果
```
注意: ScanPage函数处理某个范围的记录且返回的结果是T, 其中T是处理第一行记录时ScanRowFunc的返回结果类型. 如果没有任何记录, 即srf无法调用, 最后结果是nil

- func Get
```
func Get(name string) Clickhouse 
```
获取配置里面指定的Clickhouse实例
# Examples

- 查询demo
```
func GetPerformanceGradesAll(ctx *engine.NotifyContext) (allPerformanceGrades []*PerformanceGrade, err error) {
	psql := `select id, kungfu_id, performance_id, start_value, end_value,
			grade_name, start_ranking, end_ranking, type, season, matchCode from stat_perf_grade where type="0" `

	psrf := func(rows *sql.Rows) (interface{}, error) {
		var performanceGrade PerformanceGrade
		err = rows.Scan(&performanceGrade.Id, &performanceGrade.KungfuId, &performanceGrade.PerformanceId,
			&performanceGrade.StartValue, &performanceGrade.EndValue, &performanceGrade.GradeName, &performanceGrade.StartRanking,
			&performanceGrade.EndRanking, &performanceGrade.Type, &performanceGrade.season, &performanceGrade.matchCode)
		if err != nil {
			return nil, err
		}
		return &performanceGrade, nil
	}

	rt, err := ctx.MysqlClient.ScanAll(psql, psrf)
	if err != nil {
		return
	}
	if rt != nil {
		allPerformanceGrades = rt.([]*PerformanceGrade)
	}
	return
}
```

- 插入demo
```
	contextIpstmt := `insert into match_ctx(id, map_id, match_time, matchCode, season) values (?, ?, ?, ?, ?)`
	_, err = clichouseClient.Exec(contextIpstmt, context.Id, context.MapId, context.MatchTime, context.matchCode, context.Season)
	if err != nil {
		return
	}
```