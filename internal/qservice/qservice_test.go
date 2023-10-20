package qservice

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"kegnet.dev/qplay/internal/qprog"
	"kegnet.dev/qplay/internal/server/logger"
)

type (
	// storeMock is a mock implementation of ProgramStore.
	storeMock struct {
		saveProgramResult_Id     string
		saveProgramError         error
		saveProgramCallCount     int
		GetProgramResult_Program *qprog.Program
		GetProgramError          error
		GetProgramCallCount      int
	}

	ServiceTestSuite struct {
		suite.Suite
		Logger      *logger.Logger
		LogFn       logger.LoggingFn
		TestService Service
		storeMock   *storeMock
	}

	ErrProgramStore struct{}
)

func (e ErrProgramStore) Error() string {
	return "program store error"
}

// SaveProgram implements ProgramStore.
func (s *storeMock) SaveProgram(p *qprog.Program) (string, error) {
	s.saveProgramCallCount++
	return s.saveProgramResult_Id, s.saveProgramError
}

// GetProgram implements ProgramStore.
func (s *storeMock) GetProgram(id string) (*qprog.Program, error) {
	s.GetProgramCallCount++
	return s.GetProgramResult_Program, s.GetProgramError
}

func (s *ServiceTestSuite) SetupTest() {
	logger := logger.NewLogger(logger.LoggerOptions{
		Debug: true,
	})
	sm := &storeMock{}
	s.TestService = NewService(ServiceOptions{
		Logger: logger,
		Store:  sm,
	})

	s.Logger = logger
	s.LogFn = logger.ContextLoggingFn(&gin.Context{})
}

func (s *ServiceTestSuite) TestNewService() {
	srv := NewService(ServiceOptions{
		Logger: s.Logger,
		Store:  s.storeMock,
	})
	s.NotNil(srv)
}

func (s *ServiceTestSuite) TestSaveProgram() {
	s.storeMock = &storeMock{
		saveProgramResult_Id: "id",
	}
	pv := &ProgramValue{
		Program: qprog.Program{
			NumOfQubits: 1,
			Steps:       []qprog.Step{},
		},
	}
	id, err := s.TestService.SaveProgram(s.LogFn, pv)
	s.Nil(err)
	s.Equal("id", id)
	s.Equal(1, s.storeMock.saveProgramCallCount)
}

func (s *ServiceTestSuite) TestSaveProgramError() {
	s.storeMock = &storeMock{
		saveProgramError: new(ErrProgramStore),
	}
	pv := &ProgramValue{
		Program: qprog.Program{
			NumOfQubits: 1,
			Steps:       []qprog.Step{},
		},
	}
	id, err := s.TestService.SaveProgram(s.LogFn, pv)
	s.ErrorIs(err, new(ErrProgramStore))
	s.Equal("", id)
	s.Equal(1, s.storeMock.saveProgramCallCount)
}
