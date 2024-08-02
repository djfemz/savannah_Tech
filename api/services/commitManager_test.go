package services

import (
	"encoding/json"
	dtos "github.com/djfemz/savannahTechTask/api/dtos/responses"
	"github.com/djfemz/savannahTechTask/api/mocks"
	"github.com/djfemz/savannahTechTask/api/utils"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

var err error

func TestFetchCommitDataByTime(t *testing.T) {
	commitRepository := new(mocks.CommitRepository)
	repoMetaDataRepository := new(mocks.GithubAuxiliaryRepository)
	var commitManager = NewCommitManager(NewCommitService(commitRepository),
		NewRepoDiscoveryService(NewGithubRepoMetadataService(repoMetaDataRepository)))
	var expected *[]dtos.GitHubCommitResponse
	testTime, _ := utils.GetTimeFrom(os.Getenv("FETCH_DATE_SINCE"))
	response := utils.GetByDate(*testTime)
	log.Println("response: ", response)
	commitRepository.On("SaveAll", mock.Anything).Return(nil)
	repoMetaDataRepository.On("FindByName", mock.Anything).Return(utils.GetRepoMetaData(), nil)
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(http.MethodGet, os.Getenv("GITHUB_API_COMMIT_URL"), func(request *http.Request) (*http.Response, error) {
		res, err := httpmock.NewJsonResponse(http.StatusOK, response)
		return res, err
	})
	since, _ := time.Parse(os.Getenv("ISO_TIME_FORMAT"), os.Getenv("FETCH_DATE_SINCE"))
	data, _ := commitManager.FetchCommitDataFrom(&since)
	expectedJsonResponse, _ := json.Marshal(response)
	err = json.Unmarshal(expectedJsonResponse, &expected)
	assert.Nil(t, err)
	assert.Equal(t, data, expected)
}

func TestGetTimeFrom(t *testing.T) {
	testTime, err := utils.GetTimeFrom(os.Getenv("FETCH_DATE_SINCE"))
	assert.Nil(t, err)
	assert.NotNil(t, testTime)

}
