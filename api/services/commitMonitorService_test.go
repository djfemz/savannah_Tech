package services

import (
	"github.com/djfemz/savannahTechTask/api/utils"
	"net/http"
	"os"
	"testing"

	"github.com/djfemz/savannahTechTask/api/mocks"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFetchCommitData(t *testing.T) {
	commitRepository := new(mocks.CommitRepository)
	commitRepository.On("SaveAll", mock.Anything).Return(nil)
	commitRepository.On("FindMostRecentCommit").Return(utils.LoadTestCommits()[0], nil)
	commitMonitorService := NewCommitMonitorService(NewCommitService(commitRepository))
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(http.MethodGet, os.Getenv("GITHUB_API_COMMIT_URL"), func(request *http.Request) (*http.Response, error) {
		res, err := httpmock.NewJsonResponse(http.StatusOK, utils.LoadTestGithubCommitData())
		return res, err
	})
	data, err := commitMonitorService.FetchCommitData()
	assert.Nil(t, err)
	assert.NotEmpty(t, data)
}
