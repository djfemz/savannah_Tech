package services

import (
	"encoding/json"
	"fmt"
	"github.com/djfemz/savannahTechTask/api/utils"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/djfemz/savannahTechTask/api/appErrors"
	dtos "github.com/djfemz/savannahTechTask/api/dtos/responses"
	"github.com/djfemz/savannahTechTask/api/mappers"
)

type CommitMonitorService struct {
	*CommitService
}

func NewCommitMonitorService(commitService *CommitService) *CommitMonitorService {
	return &CommitMonitorService{
		commitService,
	}
}

func (commitMonitorService *CommitMonitorService) FetchCommitData() (githubCommitResponses *[]dtos.GitHubCommitResponse, err error) {
	commit, err := commitMonitorService.GetMostRecentCommit()
	if err != nil {
		log.Printf("[Error: %v]", err)
	}
	var resp *http.Response
	for counter := 1; counter < 100000; counter++ {
		resp, err = getData(os.Getenv("GITHUB_API_COMMIT_URL"), counter, &commit.Date)
		if err != nil {
			log.Println("error fetching data: ", err)
			return nil, err
		}
		if resp.StatusCode == http.StatusForbidden {
			time.Sleep(70 * time.Minute)
		}
	}

	githubCommitResponses, _ = extractDataInto(resp, githubCommitResponses)
	return githubCommitResponses, nil
}

func getData(url string, page int, start *time.Time) (resp *http.Response, err error) {
	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	addHeadersToRequest(req, page, start)

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	// TODO: queue up the failed requests for retries
	return resp, err
}

func (commitMonitorService *CommitMonitorService) StartJob() {
	log.Println("Repo monitoring started...")
	commitMonitorService.fetch()
}

func (commitMonitorService *CommitMonitorService) fetch() {
	data, err := commitMonitorService.FetchCommitData()
	if err != nil {
		log.Println("Error fetching commits: ", err)
	}

	if data != nil && len(*data) > 0 {
		commits := mappers.MapToCommits(data)
		err = commitMonitorService.repository.SaveAll(commits)
		if err != nil {
			log.Println("error saving commits: ", err)
		}
	}
}

func addHeadersToRequest(req *http.Request, page int, start *time.Time) {
	query := req.URL.Query()
	if start != nil {
		query.Add("since", start.String())
	}
	if page > 0 {
		query.Add("page", strconv.Itoa(page))
	}
	query.Add("Accept", utils.ACCEPT_HEADER_VALUE)
	query.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("AUTH_TOKEN")))
	req.URL.RawQuery = query.Encode()
}

func extractDataInto[t any](resp *http.Response, into *t) (*t, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, appErrors.NewCommitNotFoundError()
	}
	err := json.NewDecoder(resp.Body).Decode(&into)
	if err != nil {
		log.Println("Error reading response: ", err)
		return nil, err
	}
	return into, nil
}
