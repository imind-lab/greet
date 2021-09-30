package service

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/imind-lab/greet/mock"
	"github.com/imind-lab/greet/server/proto/greet"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type Suite struct {
	suite.Suite
	ctl    *gomock.Controller
	dmMock *mock.MockGreetDomain
	svc    GreetService
}

func (s *Suite) SetupSuite() {
	s.ctl = gomock.NewController(s.T())
	s.dmMock = mock.NewMockGreetDomain(s.ctl)
	s.svc = GreetService{
		dm: s.dmMock,
		vd: validator.New(),
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

func (s *Suite) TestGreetService_GetGreetById() {
	tests := []struct {
		name     string
		id       int32
		data     *greet.Greet
		expected *greet.GetGreetByIdResponse
	}{
		{"id-100", 100, &greet.Greet{Id: 100, Name: "koofox", ViewNum: 0, Status: 0, CreateTime: 0, CreateDatetime: "2021-06-07T08:32:34+08:00", UpdateDatetime: "2021-06-07T08:32:34+08:00"},
			&greet.GetGreetByIdResponse{
				Success: true,
				Dto:     &greet.Greet{Id: 100, Name: "koofox", ViewNum: 0, Status: 0, CreateTime: 0, CreateDatetime: "2021-06-07T08:32:34+08:00", UpdateDatetime: "2021-06-07T08:32:34+08:00"},
			},
		},
	}

	ctx := context.Background()
	for _, t := range tests {
		s.dmMock.EXPECT().GetGreetById(ctx, t.id).Return(t.data, nil)

		m, err := s.svc.GetGreetById(ctx, &greet.GetGreetByIdRequest{Id: t.id})
		require.NoError(s.T(), err)
		require.Equal(s.T(), t.expected, m)
	}
}

func (s *Suite) TestGreetService_GetGreetList() {
	tests := []struct {
		name     string
		status   int32
		lastId   int32
		pageSize int32
		page     int32
		data     *greet.GreetList
		expected *greet.GetGreetListResponse
	}{
		{"status-1", 1, 0, 3, 1,
			&greet.GreetList{
				Total:     5,
				TotalPage: 2,
				CurPage:   1,
				Datalist: []*greet.Greet{
					{Id: 100, Name: "18601038091", ViewNum: 2, Status: 1},
					{Id: 200, Name: "18601038092", ViewNum: 3, Status: 0},
					{Id: 300, Name: "18601038093", ViewNum: 4, Status: 1},
				},
			},
			&greet.GetGreetListResponse{
				Success: true,
				Data: &greet.GreetList{
					Total:     5,
					TotalPage: 2,
					CurPage:   1,
					Datalist: []*greet.Greet{
						{Id: 100, Name: "18601038091", ViewNum: 2, Status: 1},
						{Id: 200, Name: "18601038092", ViewNum: 3, Status: 0},
						{Id: 300, Name: "18601038093", ViewNum: 4, Status: 1},
					},
				},
			}},
	}

	ctx := context.Background()
	for _, t := range tests {
		s.Run(t.name, func() {
			s.dmMock.EXPECT().GetGreetList(ctx, t.status, t.lastId, t.pageSize, t.page).Return(t.data, nil)
			actual, err := s.svc.GetGreetList(ctx, &greet.GetGreetListRequest{
				Status:   t.status,
				Lastid:   t.lastId,
				Pagesize: t.pageSize,
				Page:     t.page,
			})
			require.NoError(s.T(), err)
			require.EqualValues(s.T(), t.expected, actual)
		})
	}
}
