package domain

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/imind-lab/greeter/mock"
	"github.com/imind-lab/greeter/server/model"
	"github.com/imind-lab/greeter/server/proto/greeter"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type Suite struct {
	suite.Suite
	ctl      *gomock.Controller
	repoMock *mock.MockGreeterRepository
	dm       GreeterDomain
}

func (s *Suite) SetupSuite() {
	s.ctl = gomock.NewController(s.T())
	s.repoMock = mock.NewMockGreeterRepository(s.ctl)
	s.dm = greeterDomain{
		repo: s.repoMock,
	}
}

func (s *Suite) AfterTest(_, _ string) {
}

func (s *Suite) TearDownSuite() {
	defer s.ctl.Finish()
}

func TestInit(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) TestGreeterDomain_GetGreeterById() {
	tests := []struct {
		name     string
		id       int32
		data     model.Greeter
		expected *greeter.Greeter
	}{
		{"id-100", 100, model.Greeter{Id: 100, Name: "koofox", ViewNum: 0, Status: 0, CreateTime: 0, CreateDatetime: "2021-06-07T08:32:34+08:00", UpdateDatetime: "2021-06-07T08:32:34+08:00"},
			&greeter.Greeter{Id: 100, Name: "koofox", ViewNum: 0, Status: 0, CreateTime: 0, CreateDatetime: "2021-06-07T08:32:34+08:00", UpdateDatetime: "2021-06-07T08:32:34+08:00"}},
	}

	ctx := context.Background()
	for _, t := range tests {
		s.repoMock.EXPECT().GetGreeterById(ctx, t.id).Return(t.data, nil)

		m, err := s.dm.GetGreeterById(ctx, t.id)
		require.NoError(s.T(), err)
		require.Equal(s.T(), t.expected, m)
	}
}

func (s *Suite) TestGreeterDomain_GetGreeterList() {
	tests := []struct {
		name     string
		status   int32
		lastId   int32
		pageSize int32
		page     int32
		data     []model.Greeter
		cnt      int
		expected *greeter.GreeterList
	}{
		{"status-1", 1, 0, 3, 1,
			[]model.Greeter{
				model.Greeter{Id: 100, Name: "18601038091", ViewNum: 2, Status: 1},
				model.Greeter{Id: 200, Name: "18601038092", ViewNum: 3, Status: 0},
				model.Greeter{Id: 300, Name: "18601038093", ViewNum: 4, Status: 1},
			},
			5,
			&greeter.GreeterList{
				Total:     5,
				TotalPage: 2,
				CurPage:   1,
				Datalist: []*greeter.Greeter{
					{Id: 100, Name: "18601038091", ViewNum: 2, Status: 1},
					{Id: 200, Name: "18601038092", ViewNum: 3, Status: 0},
					{Id: 300, Name: "18601038093", ViewNum: 4, Status: 1},
				},
			}},
	}

	ctx := context.Background()
	for _, t := range tests {
		s.Run(t.name, func() {
			s.repoMock.EXPECT().GetGreeterList(ctx, t.status, t.lastId, t.pageSize, t.page).Return(t.data, t.cnt, nil)
			actual, err := s.dm.GetGreeterList(ctx, t.status, t.lastId, t.pageSize, t.page)
			require.NoError(s.T(), err)
			require.EqualValues(s.T(), t.expected, actual)
		})
	}
}
