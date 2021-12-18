package persistence

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/agiledragon/gomonkey"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/imind-lab/greeter/domain/greeter/repository"
	"github.com/imind-lab/greeter/domain/greeter/repository/model"
	"github.com/imind-lab/greeter/pkg/constant"
	utilx "github.com/imind-lab/greeter/pkg/util"
	"github.com/imind-lab/micro/dao"
	redisx "github.com/imind-lab/micro/redis"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"reflect"
	"strconv"
	"testing"
	"time"
)

type Suite struct {
	suite.Suite
	mysqlDB   *gorm.DB
	mysqlMock sqlmock.Sqlmock
	redisDB   *redis.Client
	redisMock redismock.ClientMock
	repo      greeterRepository
}

func (s *Suite) SetupSuite() {
	var (
		db  *sql.DB
		err error
	)
	db, s.mysqlMock, err = sqlmock.New()
	require.NoError(s.T(), err)
	dialector := mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	})
	s.mysqlDB, err = gorm.Open(dialector, &gorm.Config{})
	require.NoError(s.T(), err)

	s.redisDB, s.redisMock = redismock.NewClientMock()
	rep := dao.NewDao(constant.DBName)
	s.repo = greeterRepository{
		Dao: rep,
	}
	s.repo.SetDBMock(s.mysqlDB)
	s.repo.SetRedisMock(s.redisDB)

}
func (s *Suite) AfterTest(_, _ string) {
	require.NoError(s.T(), s.mysqlMock.ExpectationsWereMet())
	require.NoError(s.T(), s.redisMock.ExpectationsWereMet())
}

func TestInit(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) TestGreeterRepository_CacheGreeter() {
	tests := []struct {
		name string
		data model.Greeter
	}{
		{"cache-100", model.Greeter{Id: 100, Name: "18601038091", ViewNum: 2, Status: 1}},
		{"cache-200", model.Greeter{Id: 200, Name: "18601038092", ViewNum: 3, Status: 1}},
		{"cache-300", model.Greeter{Id: 300, Name: "18601038093", ViewNum: 4, Status: 1}},
		{"cache-400", model.Greeter{Id: 400, Name: "18601038094", ViewNum: 5, Status: 0}},
	}

	ctx := context.Background()
	for _, test := range tests {
		s.Run(test.name, func() {
			key := utilx.CacheKey("greeter_", strconv.Itoa(int(test.data.Id)))
			s.redisMock.ExpectHMSet(key, redisx.FlatStruct(test.data)).SetVal(true)
			s.redisMock.ExpectExpire(key, constant.CacheMinute5).SetVal(true)
		})
		err := s.repo.CacheGreeter(ctx, test.data)
		require.NoError(s.T(), err)
	}
}

func (s *Suite) TestGreeterRepository_CreateGreeter() {
	tests := []struct {
		name string
		id   int64
		data model.Greeter
	}{
		{"cache-100", 100, model.Greeter{Name: "18601038091", ViewNum: 2, Status: 1}},
		{"cache-200", 200, model.Greeter{Name: "18601038092", ViewNum: 3, Status: 1}},
		{"cache-300", 300, model.Greeter{Name: "18601038093", ViewNum: 4, Status: 1}},
		{"cache-400", 400, model.Greeter{Name: "18601038094", ViewNum: 5, Status: 0}},
	}

	ctx := context.Background()
	for _, test := range tests {
		s.Run(test.name, func() {
			datetime := time.Now().Format("2006-01-02 15:04:05")
			s.mysqlMock.ExpectBegin()
			s.mysqlMock.ExpectExec("INSERT INTO `tbl_greeter`").WithArgs(test.data.Name, test.data.ViewNum, test.data.Status, test.data.CreateTime, datetime, datetime).WillReturnResult(sqlmock.NewResult(test.id, 1))
			s.mysqlMock.ExpectCommit()
			m, err := s.repo.CreateGreeter(ctx, test.data)
			require.NoError(s.T(), err)
			require.EqualValues(s.T(), test.id, m.Id)
			require.EqualValues(s.T(), datetime, m.CreateDatetime)
		})
	}
}

func (s *Suite) TestGreeterRepository_FindGreeterById() {
	tests := []struct {
		name     string
		rows     *sqlmock.Rows
		id       int32
		expected model.Greeter
	}{
		{"id-100",
			sqlmock.NewRows([]string{"id", "name", "view_num", "status"}).AddRow(100, "18601038090", 1, 1),
			100,
			model.Greeter{Id: 100, Name: "18601038090", ViewNum: 1, Status: 1}},
		{"id-200",
			sqlmock.NewRows([]string{"id", "name", "view_num", "status"}).AddRow(200, "18601038091", 2, 1),
			200,
			model.Greeter{Id: 200, Name: "18601038091", ViewNum: 2, Status: 1}},
	}
	for _, test := range tests {
		s.Run(test.name, func() {
			s.mysqlMock.ExpectQuery("SELECT \\* FROM `tbl_greeter`").WithArgs(test.id).WillReturnRows(test.rows)
			actual, err := s.repo.FindGreeterById(context.Background(), test.id)
			require.NoError(s.T(), err)
			require.Equal(s.T(), test.expected, actual)
		})
	}
}

func (s *Suite) TestGreeterRepository_GetGreeterById() {
	tests := []struct {
		name     string
		data     map[string]string
		val      model.Greeter
		id       int32
		opt      []repository.GreeterByIdOption
		expected model.Greeter
	}{
		{"cache-200",
			map[string]string{"id": "200", "name": "18601038091", "view_num": "2", "status": "1"},
			model.Greeter{},
			200,
			nil,
			model.Greeter{Id: 200, Name: "18601038091", ViewNum: 2, Status: 1}},
		{"cache-100",
			map[string]string{"id": "100", "name": "18601038090", "view_num": "1", "status": "1"},
			model.Greeter{},
			100,
			nil,
			model.Greeter{Id: 100, Name: "18601038090", ViewNum: 1, Status: 1}},
		{"cache-300",
			map[string]string{"id": "0"},
			model.Greeter{},
			300,
			nil,
			model.Greeter{}},
		{"cache-400",
			nil,
			model.Greeter{Id: 400, Name: "18601038099", ViewNum: 9, Status: 1},
			400,
			[]repository.GreeterByIdOption{repository.GreeterByIdRandExpire(20)},
			model.Greeter{Id: 400, Name: "18601038099", ViewNum: 9, Status: 1}},
	}

	ctx := context.Background()
	for _, test := range tests {
		s.Run(test.name, func() {
			key := utilx.CacheKey("greeter_" + strconv.Itoa(int(test.id)))
			if test.data == nil {
				s.redisMock.ExpectHGetAll(key).RedisNil()
				s.redisMock.ExpectHMSet(key, redisx.FlatStruct(test.val)).SetVal(true)
				s.redisMock.ExpectExpire(key, constant.CacheMinute5).SetVal(true)

				gomonkey.ApplyMethod(reflect.TypeOf(s.repo), "FindGreeterById", func(greeter greeterRepository, ctx context.Context, id int32) (model.Greeter, error) {
					return test.val, nil
				})
			} else {
				s.redisMock.ExpectHGetAll(key).SetVal(test.data)
			}

			actual, err := s.repo.GetGreeterById(ctx, test.id, test.opt...)

			require.NoError(s.T(), err)
			require.Equal(s.T(), test.expected, actual)
		})
	}
}

func (s *Suite) TestGreeterRepository_FindGreeterListIds() {
	tests := []struct {
		name     string
		rows     *sqlmock.Rows
		status   int32
		lastId   int32
		pageSize int32
		eptIds   []int32
		eptRedis []*redis.Z
	}{
		{"id-100",
			sqlmock.NewRows([]string{"id"}).AddRow(100).AddRow(200).AddRow(300).AddRow(400).AddRow(500),
			1,
			0,
			3,
			[]int32{100, 200, 300},
			[]*redis.Z{
				{float64(100), int32(100)},
				{float64(200), int32(200)},
				{float64(300), int32(300)},
				{float64(400), int32(400)},
				{float64(500), int32(500)}},
		},
	}

	ctx := context.Background()
	for _, test := range tests {
		s.Run(test.name, func() {
			s.mysqlMock.ExpectQuery("SELECT `id` FROM `tbl_greeter`").WithArgs(test.status).WillReturnRows(test.rows)
			ids, zs, err := s.repo.FindGreeterListIds(ctx, test.status, test.lastId, test.pageSize)
			require.NoError(s.T(), err)
			require.EqualValues(s.T(), test.eptIds, ids)
			require.EqualValues(s.T(), test.eptRedis, zs)
		})
	}
}

func (s *Suite) TestGreeterRepository_GetGreeterListIds() {
	tests := []struct {
		name     string
		status   int32
		lastId   int32
		pageSize int32
		page     int32
		data     []string
		cnt      int
		expected []int32
	}{
		{"id-100",
			1,
			0,
			3,
			1,
			[]string{"100", "200", "300"},
			5,
			[]int32{100, 200, 300},
		},
	}
	ctx := context.Background()
	for _, test := range tests {
		s.Run(test.name, func() {
			start := (test.page - 1) * test.pageSize
			key := utilx.CacheKey("greeter_ids_", strconv.Itoa(int(test.status)))
			s.redisMock.ExpectType(key).SetVal("zset")
			s.redisMock.ExpectZCard(key).SetVal(int64(test.cnt))
			s.redisMock.ExpectZRevRangeByScore(key, &redis.ZRangeBy{Max: "+inf", Min: "-inf", Offset: int64(start), Count: int64(test.pageSize)}).SetVal(test.data)
			ids, i, err := s.repo.GetGreeterListIds(ctx, test.status, test.lastId, test.pageSize, test.page)
			require.NoError(s.T(), err)
			require.EqualValues(s.T(), test.cnt, i)
			require.EqualValues(s.T(), test.expected, ids)
		})
	}
}
