package valuer

import (
	"testing"

	"database/sql"

	"database/sql/driver"

	"github.com/startdusk/go-libs/orm/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// go test -bench=BenchmarkSetColumns -benchtime=10000x -benchmem
// 执行 10000 次, 输出内存分配
// 结果应该是 unsfae 比 reflect快(reflect封装了unsafe, 所以比unsafe慢)

func BenchmarkSetColumns(b *testing.B) {
	// 不是特别准, 因为我们使用sqlmock模拟了数据库(这个耗时不固定)
	b.Run("reflect", func(b *testing.B) {
		benchmarkSetColumns(b, NewReflectValue)
	})

	b.Run("unsafe", func(b *testing.B) {
		benchmarkSetColumns(b, NewUnsafeValue)
	})
}

func benchmarkSetColumns(b *testing.B, creator Creator) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(b, err)
	defer mockDB.Close()
	rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
	// 我们benchmark测试b.N次, 也就是需要b.N行数据
	row := []driver.Value{"1", "Tom", "18", "Jerry"}
	for i := 0; i < b.N; i++ {
		rows.AddRow(row...)
	}
	mock.ExpectQuery("SELECT xxx").WillReturnRows(rows)
	newRows, err := mockDB.Query("SELECT xxx")
	require.NoError(b, err)

	// 重置定时(因为设置数据库操作 不在 benchmark 测试的函数里面)
	b.ResetTimer()

	r := model.NewRegistry()
	m, err := r.Get(&TestModel{})
	require.NoError(b, err)
	for i := 0; i < b.N; i++ {
		newRows.Next()
		val := creator(m, &TestModel{})
		// 基准测试不需要比较最终结果, 最终结果在单元测试中比较
		_ = val.SetColumns(newRows)
	}
}

func testSetColumns(t *testing.T, creator Creator) {
	cases := []struct {
		name string

		// entity 一定是指针
		entity any

		rows       func() *sqlmock.Rows
		wantErr    error
		wantEntity any
	}{
		{
			name:   "测试sql query全部字段, 按顺序, 并给结构体指针赋值",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "age", "last_name"})
				rows.AddRow("1", "Tom", "18", "Jerry")
				return rows
			},
			wantEntity: &TestModel{
				ID:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			},
		},
		{
			name:   "测试sql query全部字段, 不按顺序, 并给结构体指针赋值",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"})
				rows.AddRow("1", "Tom", "Jerry", "18")
				return rows
			},
			wantEntity: &TestModel{
				ID:        1,
				FirstName: "Tom",
				Age:       18,
				LastName:  &sql.NullString{Valid: true, String: "Jerry"},
			},
		},
		{
			name:   "测试sql query部分字段",
			entity: &TestModel{},
			rows: func() *sqlmock.Rows {
				rows := sqlmock.NewRows([]string{"id", "first_name"})
				rows.AddRow("1", "Tom")
				return rows
			},
			wantEntity: &TestModel{
				ID:        1,
				FirstName: "Tom",
			},
		},
	}

	r := model.NewRegistry()
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// 转换 rows
			mock.ExpectQuery("SELECT xxx").WillReturnRows(c.rows())
			rows, err := mockDB.Query("SELECT xxx")
			require.NoError(t, err)

			rows.Next()

			m, err := r.Get(c.entity)
			require.NoError(t, err)
			val := creator(m, c.entity)

			err = val.SetColumns(rows)
			assert.Equal(t, c.wantErr, err)
			if err != nil {
				return
			}

			// 检测是否已经修改了
			assert.Equal(t, c.wantEntity, c.entity)
		})
	}
}

type TestModel struct {
	ID        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
