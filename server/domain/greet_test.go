package domain

import (
	"context"
	"github.com/imind-lab/greet/mock"
	"github.com/imind-lab/greet/server/model"
	"github.com/imind-lab/greet/server/proto/greet"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type Suite struct {
	suite.Suite
	ctl      *gomock.Controller
	repoMock *mock.MockGreetRepository
	dm       GreetDomain
}

func (s *Suite) SetupSuite() {
	s.ctl = gomock.NewController(s.T())
	s.repoMock = mock.NewMockGreetRepository(s.ctl)
	s.dm = greetDomain{
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

func (s *Suite) TestGreetDomain_GetGreetById() {
	tests := []struct {
		name     string
		id       int32
		data     model.Greet
		expected *greet.Greet
	}{
		{"id-100", 100, model.Greet{Id: 100, Name: "koofox", ViewNum: 0, Status: 0, CreateTime: 0, CreateDatetime: "2021-06-07T08:32:34+08:00", UpdateDatetime: "2021-06-07T08:32:34+08:00"},
			&greet.Greet{Id: 100, Name: "koofox", ViewNum: 0, Status: 0, CreateTime: 0, CreateDatetime: "2021-06-07T08:32:34+08:00", UpdateDatetime: "2021-06-07T08:32:34+08:00"}},
	}

	ctx := context.Background()
	for _, t := range tests {
		s.repoMock.EXPECT().GetGreetById(ctx, t.id).Return(t.data, nil)

		m, err := s.dm.GetGreetById(ctx, t.id)
		require.NoError(s.T(), err)
		require.Equal(s.T(), t.expected, m)
	}
}

func (s *Suite) TestGreetDomain_GetGreetList() {
	tests := []struct {
		name     string
		status   int32
		lastId   int32
		pageSize int32
		page     int32
		data     []model.Greet
		cnt      int
		expected *greet.GreetList
	}{
		{"status-1", 1, 0, 3, 1,
			[]model.Greet{
				model.Greet{Id: 100, Name: "18601038091", ViewNum: 2, Status: 1},
				model.Greet{Id: 200, Name: "18601038092", ViewNum: 3, Status: 0},
				model.Greet{Id: 300, Name: "18601038093", ViewNum: 4, Status: 1},
			},
			5,
			&greet.GreetList{
				Total:     5,
				TotalPage: 2,
				CurPage:   1,
				Datalist: []*greet.Greet{
					{Id: 100, Name: "18601038091", ViewNum: 2, Status: 1},
					{Id: 200, Name: "18601038092", ViewNum: 3, Status: 0},
					{Id: 300, Name: "18601038093", ViewNum: 4, Status: 1},
				},
			}},
	}

	ctx := context.Background()
	for _, t := range tests {
		s.Run(t.name, func() {
			s.repoMock.EXPECT().GetGreetList(ctx, t.status, t.lastId, t.pageSize, t.page).Return(t.data, t.cnt, nil)
			actual, err := s.dm.GetGreetList(ctx, t.status, t.lastId, t.pageSize, t.page)
			require.NoError(s.T(), err)
			require.EqualValues(s.T(), t.expected, actual)
		})
	}
}
