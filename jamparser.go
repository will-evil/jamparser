package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"jamparser/pkg/csv"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const siteUrl = "https://www.jamesedition.com/"

type Makes map[string]string

type Model struct {
	Name string `json:"n"`
	Slug string `json:"i"`
}

type Models []Model

var httpClient *http.Client

func init() {
	httpClient = &http.Client{
		Timeout: time.Second * 15,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func main() {
	path := flag.String("path", "result.csv", "path to file for save data")
	target := flag.String("target", "", "parsing target (cars, motorcycles, e.t.c)")

	flag.Parse()

	if *target == "" {
		fmt.Println("you must provide target (-target=cars)")
		return
	}

	makes, err := getMakes(*target)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(fmt.Sprintf("found %d makes", len(makes)))

	var data = csv.CSVData{{"Model name", "Model Slug", "Make Name", "Make Slug"}}
	for slug, name := range makes {
		fmt.Println(fmt.Sprintf("processing make '%s'", name))

		models, err := getModels(slug, *target)
		if err != nil {
			fmt.Println(fmt.Sprintf("get models for make '%s' is failed", name))
		}

		fmt.Println(fmt.Sprintf("found %d models for '%s'", len(models), name))
		for _, model := range models {
			data = append(data, []string{
				model.Name,
				model.Slug,
				name,
				slug,
			})
		}

		fmt.Println(fmt.Sprintf("make %s was processed", name))
	}

	err = csv.Write(*path, data)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(fmt.Sprintf("data writed to %s", *path))
}

func getMakes(target string) (Makes, error) {
	makes := make(Makes)

	body, err := getRequest(siteUrl + target)
	if err != nil {
		return makes, err
	}

	re := regexp.MustCompile(`(?U)id="brand".*>[\s\S]*<\/select>`)
	re2 := regexp.MustCompile(`<option.*value="(.+)".*>(.+)<`)
	if err != nil {
		return makes, err
	}

	if re.MatchString(string(body)) {
		results := re2.FindAllSubmatch(re.Find(body), -1)
		for _, res := range results {
			makes[strings.TrimSpace(string(res[1]))] = strings.TrimSpace(string(res[2]))
		}
	} else {
		return makes, errors.New("can not parse page for get makes")
	}

	return makes, nil
}

func getModels(makeName string, target string) (Models, error) {
	var models Models

	body, err := getRequest(fmt.Sprintf("%smodels?category=%s&brand=%s", siteUrl, target, makeName))
	if err != nil {
		return models, err
	}

	err = json.Unmarshal(body, &models)
	if err != nil {
		return models, err
	}

	return models, nil
}

func getRequest(url string) ([]byte, error) {
	var body []byte
	var err error

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Request Failed. Status Code %d", res.StatusCode)
	}

	return body, nil
}
