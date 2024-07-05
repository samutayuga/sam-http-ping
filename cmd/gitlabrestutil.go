package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	personalToken string
	gitlabRepo    ProjectRepoList
	msgTemplate   *template.Template
	appConfig     AppConfig
)

type GitlabRepo struct {
	Id              int
	Name            string
	Url             string
	Artefacts       []string
	ReleaseEndPoint string
}

func (g *GitlabRepo) makeRepo(projectInMap map[string]interface{}) {
	g.Id = int(projectInMap["id"].(float64))
	g.Name = projectInMap["name"].(string)
	g.Url = projectInMap["web_url"].(string)
	g.ReleaseEndPoint = fmt.Sprintf("%s/projects/%d/releases", appConfig.GitlabSearch.BaseUrl, g.Id)
}

type GitlabGroup struct {
	Name        string
	Id          int
	Url         string
	GitlabRepos []*GitlabRepo
	Endpoint    string
}
type GitlabRelease struct {
	Tag  string
	Name string
}
type GitlabGroupManager interface {
	Make(grInMap map[string]interface{})
}

type ProjectRepoList struct {
	httpClient      *http.Client
	GitBaseEndPoint string
	GroupIds        []int
	GitlabGroups    []*GitlabGroup
}

type ProjectRepoCURD interface {
	Parse()
	AmmendGId(configuredGroups []GitlabGroupDescriptor)
}

func handleHttpGet(url string, httpClient *http.Client) []map[string]interface{} {

	var anyError error
	var gLabModules []map[string]interface{}
	var resp *http.Response
	var errGetGroup error
	var respBody []byte
	var gitlabGroupReq *http.Request

	if gitlabGroupReq, anyError = http.NewRequest("GET", url, nil); anyError != nil {
		Logger.Fatal("error while creating the http request", zap.String("url", url), zap.Error(anyError))
	}
	gitlabGroupReq.Header.Add("PRIVATE-TOKEN", personalToken)
	if resp, errGetGroup = httpClient.Do(gitlabGroupReq); errGetGroup != nil {
		Logger.Fatal("error while contacting end point", zap.String("url", url), zap.Error(errGetGroup))
	}
	defer resp.Body.Close()
	if respBody, anyError = io.ReadAll(resp.Body); anyError != nil {
		Logger.Fatal("Error while reading the response body", zap.Error(anyError))
	}

	if anyError = json.Unmarshal(respBody, &gLabModules); anyError != nil {
		Logger.Fatal("Error while unmarshalling body", zap.Error(anyError))
	}
	return gLabModules
}
func (p *ProjectRepoList) Parse() {
	groupUrls := fmt.Sprintf("%s/%s?page=%d&per_page=%d", appConfig.GitlabSearch.BaseUrl, "groups", appConfig.GitlabSearch.Page, appConfig.GitlabSearch.RowPerPage)
	gLabModules := handleHttpGet(groupUrls, gitlabRepo.httpClient)
	for _, v := range gLabModules {
		gitGroup := GitlabGroup{}
		gitGroup.Make(v)
		p.GitlabGroups = append(p.GitlabGroups, &gitGroup)
		p.GroupIds = append(p.GroupIds, gitGroup.Id)
	}
}

func (g *GitlabGroup) Make(grInMap map[string]interface{}) {
	g.Id = int(grInMap["id"].(float64))
	g.Name = grInMap["name"].(string)
	g.Url = grInMap["web_url"].(string)
	g.Endpoint = fmt.Sprintf("%s/groups/%d/projects?page=%d&per_page=%d", appConfig.GitlabSearch.BaseUrl, g.Id, appConfig.GitlabSearch.Page, appConfig.GitlabSearch.RowPerPage)
}
func (p *ProjectRepoList) AmmendGId(configuredGroups []GitlabGroupDescriptor) {
	for _, v := range configuredGroups {
		gitGroup := GitlabGroup{Id: v.Id, Name: v.Name, Endpoint: fmt.Sprintf("%s/groups/%d/projects?page=%d&per_page=%d", appConfig.GitlabSearch.BaseUrl, v.Id, appConfig.GitlabSearch.Page, appConfig.GitlabSearch.RowPerPage)}
		p.GitlabGroups = append(p.GitlabGroups, &gitGroup)
	}
}
func init() {
	var err error
	if msgTemplate, err = template.ParseGlob("*.gotmpl"); err != nil {
		log.Fatalf("error while parsing the template %v\n", err)
	} else {
		log.Printf("template %s to display is initialized\n", msgTemplate.Name())
	}
}

var gitlabRepListCmd = &cobra.Command{
	Use:   "gitlabcaptor",
	Short: "The utility to get the gitlab project list",
	Long: `This is used to retrieve all the gitlab project under a certain group
	`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		viper.SetConfigFile("config/app-config.yaml")
		if errread := viper.ReadInConfig(); errread != nil {
			Logger.Fatal("Error while reading config", zap.Error(errread))
		}
		Logger.Info("searching for", zap.Any("groups", viper.AllKeys()))

		appConfig = AppConfig{EndPoints: make([]map[string]interface{}, 0), GitlabSearch: GitLabSearchInfo{}}
		if unmErr := viper.Unmarshal(&appConfig); unmErr != nil {
			Logger.Fatal("error while unmarshalling config", zap.Error(unmErr))
		}
		Logger.Info("Config is", zap.Any("content", appConfig))

		tr := &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		}
		httpClient := &http.Client{
			Transport: tr,
		}
		gitlabRepo = ProjectRepoList{httpClient: httpClient, GitBaseEndPoint: appConfig.GitlabSearch.BaseUrl, GroupIds: make([]int, 0), GitlabGroups: make([]*GitlabGroup, 0)}
		gitlabRepo.AmmendGId(appConfig.GitlabSearch.GroupDescriptor.GroupByIds)

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		groupUrls := fmt.Sprintf("%s/%s?page=%d&per_page=%d", appConfig.GitlabSearch.BaseUrl, "groups", appConfig.GitlabSearch.Page, appConfig.GitlabSearch.RowPerPage)
		bodyInList := handleHttpGet(groupUrls, gitlabRepo.httpClient)
		for _, grpInMap := range bodyInList {
			currentWebUrl := grpInMap["web_url"].(string)

			for _, s := range appConfig.GitlabSearch.GroupDescriptor.GroupByUrls {
				if s == currentWebUrl {
					gitlabGr := GitlabGroup{GitlabRepos: make([]*GitlabRepo, 0)}
					gitlabGr.Make(grpInMap)
					gitlabRepo.GitlabGroups = append(gitlabRepo.GitlabGroups, &gitlabGr)
					break
				}
			}

		}
		Logger.Info("Complete gathering the gitlab groups", zap.Int("count", len(gitlabRepo.GitlabGroups)))
		for _, gGroup := range gitlabRepo.GitlabGroups {
			bodyInList = handleHttpGet(gGroup.Endpoint, gitlabRepo.httpClient)
			for _, c := range bodyInList {
				gitLabProject := GitlabRepo{}
				gitLabProject.makeRepo(c)
				gGroup.GitlabRepos = append(gGroup.GitlabRepos, &gitLabProject)

				//get the releases
				//bodyInList = handleHttpGet(gitLabProject.ReleaseEndPoint)

			}
		}

		if errDisplay := msgTemplate.ExecuteTemplate(os.Stdout, "displayer.gotmpl", gitlabRepo); errDisplay != nil {
			Logger.Fatal("error while writing into template", zap.Error(errDisplay))
		}

		return nil
	},
}
