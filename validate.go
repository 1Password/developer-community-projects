package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

type Project struct {
	Category    string   `json:"category" validate:"required,oneof=article repo video"`
	Id          string   `json:"id" validate:"required,unique_id,alphanumeric_dashes"`
	Title       string   `json:"title" validate:"required,plaintext"`
	Author      string   `json:"author" validate:"required,plaintext"`
	Url         string   `json:"url" validate:"required,url,stable_url"`
	Description string   `json:"description" validate:"omitempty,plaintext"`
	Date        string   `json:"date" validate:"required,datetime=2006-01-02"`
	Tags        []string `json:"tags" validate:"required,dive,plaintext"`
}

func main() {
	var validatedProjects []Project

	currentBranch := getEnvVar("CURRENT_BRANCH")
	fmt.Println("Current branch:", currentBranch)

	branchData, err := ioutil.ReadFile("projects.json")
	if err != nil {
		fmt.Println("Error reading first file:", err)
		os.Exit(1)
	}

	branchProjects := projectsFromJson(branchData)

	if currentBranch == "main" {
		validatedProjects = branchProjects
	} else {
		mainData := getEnvVar("MAIN_PROJECTS_DATA")
		mainProjects := projectsFromJson([]byte(mainData))

		for i, projectOnCurrentBranch := range branchProjects {
			projectUpdated := i < len(mainProjects) && !reflect.DeepEqual(projectOnCurrentBranch, mainProjects[i])
			projectNewlyAdded := i >= len(mainProjects)

			if projectUpdated || projectNewlyAdded {
				validatedProjects = append(validatedProjects, projectOnCurrentBranch)
			}
		}
	}

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	validationFailed := false

	validate := validator.New()
	validate.RegisterValidation("unique_id", UniqueId)
	validate.RegisterValidation("alphanumeric_dashes", AlphaNumDashes)
	validate.RegisterValidation("plaintext", func(fl validator.FieldLevel) bool {
		return PrintOnly(fl) && NoEmojis(fl) && NoUrls(fl) && NoHtmlChars(fl)
	})
	validate.RegisterValidation("stable_url", func(fl validator.FieldLevel) bool {
		return StableUrl(fl, client)
	})

	validatingJSON, _ := json.MarshalIndent(validatedProjects, "", "	")
	fmt.Println("Validating projects:\n", string(validatingJSON))

	for i, p := range validatedProjects {
		err = validate.Struct(p)
		if err != nil {
			fmt.Printf("Error validating project %d: %s\n", i+1, err)
			validationFailed = true
		}
	}

	if validationFailed {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func projectsFromJson(data []byte) []Project {
	var projects []Project
	err := json.Unmarshal([]byte(data), &projects)
	if err != nil {
		fmt.Println("Error unmarshaling data as JSON:", err)
		os.Exit(1)
	}

	return projects
}

func getEnvVar(name string) string {
	value, ok := os.LookupEnv(name)
	if !ok || len(value) == 0 {
		fmt.Printf("%s not set\n", name)
		os.Exit(1)
	}

	return value
}

var ids = map[string]bool{}

func UniqueId(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if ids[value] {
		return false
	}
	ids[value] = true
	return true
}

func PrintOnly(fl validator.FieldLevel) bool {
	for _, r := range fl.Field().String() {
		if !unicode.IsPrint(r) {
			return false
		}
	}

	return true
}

func AlphaNumDashes(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	match, _ := regexp.MatchString("^[a-zA-Z0-9-]+$", value)

	return match
}

func NoEmojis(fl validator.FieldLevel) bool {
	for _, r := range fl.Field().String() {
		if r >= 0x1F600 && r <= 0x1F64F {
			return false
		}
	}

	return true
}

func NoUrls(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	re := regexp.MustCompile(`https?://\S+`)

	return !re.MatchString(value)
}

func NoHtmlChars(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Keeping this pretty loose as it's not uncommon
	// for titles to have ampersands and quotes in them,
	// and the client is going to encode it anyhow
	re := regexp.MustCompile(`[<|>]`)

	return !re.MatchString(value)
}

func StableUrl(fl validator.FieldLevel, client http.Client) bool {
	value := fl.Field().String()
	res, err := client.Get(value)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	// Medium articles with a custom domain do a redirect through
	// medium.com, so this is a special case to allow the 307
	if res.StatusCode == 307 {
		url, err := url.Parse(res.Header.Get("Location"))
		if err != nil {
			return false
		}
		if url.Host == "medium.com" {
			return true
		}
	}

	pass := res.StatusCode == 200

	if !pass {
		fmt.Printf("Received status code %d from %s\n", res.StatusCode, value)
	}

	return pass
}
