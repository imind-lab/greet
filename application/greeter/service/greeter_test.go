package service

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/imind-lab/greeter/application/greeter/proto"
	"github.com/imind-lab/greeter/test/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

type Suite struct {
	suite.Suite
	ctl    *gomock.Controller
	dmMock *mock.MockGreeterDomain
	svc    GreeterService
}

func (s *Suite) SetupSuite() {
	s.ctl = gomock.NewController(s.T())
	s.dmMock = mock.NewMockGreeterDomain(s.ctl)
	s.svc = GreeterService{
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

func (s *Suite) TestGreeterService_GetGreeterById() {
	tests := []struct {
		name     string
		id       int32
		data     *greeter.Greeter
		expected *greeter.GetGreeterByIdResponse
	}{
		{"id-100", 100, &greeter.Greeter{Id: 100, Name: "koofox", ViewNum: 0, Status: 0, CreateTime: 0, CreateDatetime: "2021-06-07T08:32:34+08:00", UpdateDatetime: "2021-06-07T08:32:34+08:00"},
			&greeter.GetGreeterByIdResponse{
				Success: true,
				Dto:     &greeter.Greeter{Id: 100, Name: "koofox", ViewNum: 0, Status: 0, CreateTime: 0, CreateDatetime: "2021-06-07T08:32:34+08:00", UpdateDatetime: "2021-06-07T08:32:34+08:00"},
			},
		},
	}

	ctx := context.Background()
	for _, t := range tests {
		s.dmMock.EXPECT().GetGreeterById(ctx, t.id).Return(t.data, nil)

		m, err := s.svc.GetGreeterById(ctx, &greeter.GetGreeterByIdRequest{Id: t.id})
		require.NoError(s.T(), err)
		require.Equal(s.T(), t.expected, m)
	}
}

func (s *Suite) TestGreeterService_GetGreeterList() {
	tests := []struct {
		name     string
		status   int32
		lastId   int32
		pageSize int32
		page     int32
		data     *greeter.GreeterList
		expected *greeter.GetGreeterListResponse
	}{
		{"status-1", 1, 0, 3, 1,
			&greeter.GreeterList{
				Total:     5,
				TotalPage: 2,
				CurPage:   1,
				Datalist: []*greeter.Greeter{
					{Id: 100, Name: "18601038091", ViewNum: 2, Status: 1},
					{Id: 200, Name: "18601038092", ViewNum: 3, Status: 0},
					{Id: 300, Name: "18601038093", ViewNum: 4, Status: 1},
				},
			},
			&greeter.GetGreeterListResponse{
				Success: true,
				Data: &greeter.GreeterList{
					Total:     5,
					TotalPage: 2,
					CurPage:   1,
					Datalist: []*greeter.Greeter{
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
			s.dmMock.EXPECT().GetGreeterList(ctx, t.status, t.lastId, t.pageSize, t.page).Return(t.data, nil)
			actual, err := s.svc.GetGreeterList(ctx, &greeter.GetGreeterListRequest{
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
