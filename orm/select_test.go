package orm

import (
	"database/sql"
	"testing"

	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/startdusk/go-libs/orm/internal/errs"
	"github.com/startdusk/go-libs/orm/internal/valuer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"errors"
	_ "github.com/mattn/go-sqlite3"
)

func Test_Select_Join(t *testing.T) {
	db := memoryDB(t)
	type Order struct {
		ID        int
		UsingCol1 string
		UsingCol2 string
	}
	type OrderDetail struct {
		OrderID int
		ItemID  int

		UsingCol1 string
		UsingCol2 string
	}
	type Item struct {
		ID int
	}

	cases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "specify table",
			q:    NewSelector[Order](db).From(TableOf(&OrderDetail{})),
			wantQuery: &Query{
				SQL: "SELECT * FROM `order_detail`;",
			},
		},
		{
			name: "join using",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{})
				t2 := TableOf(&OrderDetail{})
				t3 := t1.Join(t2).Using("UsingCol1", "UsingCol2")
				return NewSelector[Order](db).From(t3)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` JOIN `order_detail` USING (`using_col_1`,`using_col_2`));",
			},
		},
		{
			name: "join on",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := t1.Join(t2).On(t1.C("ID").Eq(t2.C("OrderID")))
				return NewSelector[Order](db).From(t3)
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` AS `t1` JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`);",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			query, err := c.q.Build()
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantQuery, query)
		})
	}
}

func Test_Selector_Select(t *testing.T) {
	db := memoryDB(t)
	cases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "multiple columns",
			q:    NewSelector[TestModel](db).Select(C("FirstName"), C("LastName")),
			wantQuery: &Query{
				SQL: "SELECT `first_name`, `last_name` FROM `test_model`;",
			},
		},
		{
			name:    "Invalid",
			q:       NewSelector[TestModel](db).Select(C("Invalid")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			name:    "AVG Invalid",
			q:       NewSelector[TestModel](db).Select(Avg("Invalid")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
		{
			name: "AVG",
			q:    NewSelector[TestModel](db).Select(Avg("Age")),
			wantQuery: &Query{
				SQL: "SELECT AVG(`age`) FROM `test_model`;",
			},
		},
		{
			name: "SUM",
			q:    NewSelector[TestModel](db).Select(Sum("Age")),
			wantQuery: &Query{
				SQL: "SELECT SUM(`age`) FROM `test_model`;",
			},
		},
		{
			name: "COUNT",
			q:    NewSelector[TestModel](db).Select(Count("Age")),
			wantQuery: &Query{
				SQL: "SELECT COUNT(`age`) FROM `test_model`;",
			},
		},
		{
			name: "MAX",
			q:    NewSelector[TestModel](db).Select(Max("Age")),
			wantQuery: &Query{
				SQL: "SELECT MAX(`age`) FROM `test_model`;",
			},
		},
		{
			name: "MIN",
			q:    NewSelector[TestModel](db).Select(Min("Age")),
			wantQuery: &Query{
				SQL: "SELECT MIN(`age`) FROM `test_model`;",
			},
		},
		{
			name: "multiple aggregate",
			q:    NewSelector[TestModel](db).Select(Max("Age"), Min("Age")),
			wantQuery: &Query{
				SQL: "SELECT MAX(`age`), MIN(`age`) FROM `test_model`;",
			},
		},
		{
			name: "raw expression",
			q:    NewSelector[TestModel](db).Select(Raw("COUNT(DISTINCT `age`)")),
			wantQuery: &Query{
				SQL: "SELECT COUNT(DISTINCT `age`) FROM `test_model`;",
			},
		},
		{
			name: "columns alias",
			q:    NewSelector[TestModel](db).Select(C("FirstName").As("my_name")),
			wantQuery: &Query{
				SQL: "SELECT `first_name` AS `my_name` FROM `test_model`;",
			},
		},
		{
			name: "aggregate alias",
			q:    NewSelector[TestModel](db).Select(Max("Age").As("my_age")),
			wantQuery: &Query{
				SQL: "SELECT MAX(`age`) AS `my_age` FROM `test_model`;",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			query, err := c.q.Build()
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantQuery, query)
		})
	}
}

func Test_Selector_Build(t *testing.T) {
	db := memoryDB(t)
	cases := []struct {
		name    string
		builder QueryBuilder

		wantQuery *Query
		wantErr   error
	}{
		{
			name:    "select_empty_where",
			builder: NewSelector[TestModel](db).Where(),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "select_no_from",
			builder: NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model`;",
				Args: nil,
			},
			wantErr: nil,
		},
		{
			name:    "select_from_test_model_where_age=18",
			builder: NewSelector[TestModel](db).Where(C(`Age`).Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` = ?;",
				Args: []any{18},
			},
			wantErr: nil,
		},

		{
			name:    "select_from_test_model_where_not(age=18)",
			builder: NewSelector[TestModel](db).Where(Not(C(`Age`).Eq(18))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE  NOT (`age` = ?);",
				Args: []any{18},
			},
			wantErr: nil,
		},

		{
			name:    "select_from_test_model_where_age=18_and_first_name=tom",
			builder: NewSelector[TestModel](db).Where(C(`Age`).Eq(18).And(C("FirstName").Eq("tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) AND (`first_name` = ?);",
				Args: []any{18, "tom"},
			},
			wantErr: nil,
		},

		{
			name:    "select_from_test_model_where_age=18_or_first_name=tom",
			builder: NewSelector[TestModel](db).Where(C(`Age`).Eq(18).Or(C("FirstName").Eq("tom"))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` = ?) OR (`first_name` = ?);",
				Args: []any{18, "tom"},
			},
			wantErr: nil,
		},
		{
			name:    "invalid column",
			builder: NewSelector[TestModel](db).Where(C(`Age`).Eq(18).Or(C("XXX").Eq("tom"))),

			wantErr: errs.NewErrUnknownField("XXX"),
		},
		{
			name:    "raw expression as predicate",
			builder: NewSelector[TestModel](db).Where(Raw("`age`<?", 18).AsPredicate()),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age`<?);",
				Args: []any{18},
			},
		},
		{
			name:    "raw expression used in predicate",
			builder: NewSelector[TestModel](db).Where(C("ID").Eq(Raw("`age`+?", 1))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` = (`age`+?);",
				Args: []any{1},
			},
		},
		{
			name:    "column alias in where", // where 部分的字段是不允许使用 AS
			builder: NewSelector[TestModel](db).Where(C("ID").As("my_id").Eq(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `id` = ?;",
				Args: []any{18},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			q, err := c.builder.Build()
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantQuery, q)
		})
	}
}

// TODO: 待测试这个
func Test_Selector_GetMulit(t *testing.T) {}

func Test_Selector_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	db, err := OpenDB(mockDB, DBWithDialect(DialectMySQL))
	assert.NoError(t, err)

	// 对应 query error
	queryError := errors.New("query error")
	mock.ExpectQuery("SELECT .*").WillReturnError(queryError)

	// 对应 no rows
	rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	// // 对应 scan error
	// rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	// // 本来ID应该应该是数字类型, 但故意给了个abc, mock scan error
	// rows.AddRow("abc", "Tom", "18", "Jerry")
	// mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	// 对应 query row success
	rows = sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	// 数据库查询出来的数据返回的都是文本类型, 所以这里可以用字符串
	rows.AddRow("1", "Tom", "18", "Jerry")
	mock.ExpectQuery("SELECT .*").WillReturnRows(rows)

	cases := []struct {
		name    string
		s       *Selector[TestModel]
		wantErr error
		wantRes *TestModel
	}{
		{
			name:    "invalid query",
			s:       NewSelector[TestModel](db).Where(C("xxx").Eq(1)),
			wantErr: errs.NewErrUnknownField("xxx"),
		},
		{
			name:    "query error",
			s:       NewSelector[TestModel](db).Where(C("ID").Eq(1)),
			wantErr: queryError,
		},
		{
			name:    "no rows",
			s:       NewSelector[TestModel](db).Where(C("ID").Lt(1)),
			wantErr: ErrNoRows,
		},
		// {
		// 	name:    "scan error",
		// 	s:       NewSelector[TestModel](db).Where(C("ID").Lt(1)),
		// 	wantErr: errors.New(""), // 很难构造这个 rows.Scan 的错误
		// },
		{
			name: "data",
			s:    NewSelector[TestModel](db).Where(C("ID").Eq(1)),
			wantRes: &TestModel{
				ID:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := c.s.Get(context.Background())
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, c.wantRes, res)
		})
	}
}

type TestModel struct {
	ID        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}

func (t TestModel) CreateSQL() string {
	return `
		CREATE TABLE IF NOT EXISTS test_model (
			id INTEGER PRIMARY KEY,
			first_name TEXT NOT NULL,
			age INTEGER,
			last_name TEXT NOT NULL
		)
	`
}

func memoryDB(t *testing.T, opts ...DBOption) *DB {
	db, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory", opts...)
	require.NoError(t, err)
	return db
}

// go test -bench=BenchmarkQueriesGet -benchtime=10000x -benchmem
func BenchmarkQueriesGet(b *testing.B) {
	db, err := Open("sqlite3", "file:benchmark_get.db?cache=shared&mode=memory")
	if err != nil {
		b.Fatal(err)
	}
	testModel := TestModel{}
	if _, err := db.db.ExecContext(context.Background(), testModel.CreateSQL()); err != nil {
		b.Fatal(err)
	}

	res, err := db.db.Exec("INSERT INTO `test_model`(`id`, `first_name`, `age`, `last_name`) VALUES (?, ?, ?, ?)", 12, "Tom", 18, "Jerry")
	if err != nil {
		b.Fatal(err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		b.Fatal(err)
	}
	if affected == 0 {
		b.Fatal()
	}

	b.Run("unsafe", func(b *testing.B) {
		db.creator = valuer.NewUnsafeValue
		for i := 0; i < b.N; i++ {
			if _, err := NewSelector[TestModel](db).Get(context.Background()); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("reflect", func(b *testing.B) {
		db.creator = valuer.NewReflectValue
		for i := 0; i < b.N; i++ {
			if _, err := NewSelector[TestModel](db).Get(context.Background()); err != nil {
				b.Fatal(err)
			}
		}
	})
}
