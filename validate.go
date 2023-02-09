package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"unicode/utf8"

	"github.com/go-playground/validator"
)

type Project struct {
	Type        string   `json:"type" validate:"required,oneof=article repo video"`
	Id          string   `json:"id" validate:"required,unique_id,alphanumeric_dashes"`
	Title       string   `json:"title" validate:"required,no_emojis,no_urls"`
	Author      string   `json:"author" validate:"required,no_emojis,no_urls"`
	Url         string   `json:"url" validate:"required,url,no_redirect"`
	Description string   `json:"description" validate:"omitempty,no_emojis,no_urls"`
	Date        string   `json:"date" validate:"omitempty,date_for_article_video"`
	Tags        []string `json:"tags" validate:"required,dive,no_emojis,no_urls"`
}

type changedObject struct {
	Index int
	Value Project
}

func main() {
	var validatedProjects []Project

	currentBranch, ok := os.LookupEnv("CURRENT_BRANCH")
	if !ok || len(currentBranch) == 0 {
		fmt.Printf("%s not set\n", "CURRENT_BRANCH")
	}

	fmt.Println("Current branch:", currentBranch)

	branchData, err := ioutil.ReadFile("projects.json")
	if err != nil {
		fmt.Println("Error reading first file:", err)
		os.Exit(1)
	}

	var branchProjects []Project
	err = json.Unmarshal(branchData, &branchProjects)
	if err != nil {
		fmt.Println("Error unmarshaling branchData as JSON:", err)
		os.Exit(1)
	}

	if currentBranch == "main" {
		validatedProjects = branchProjects
	} else {
		mainData, ok := os.LookupEnv("MAIN_PROJECTS_DATA")
		if !ok || len(mainData) == 0 {
			fmt.Printf("%s not set\n", "MAIN_PROJECTS_DATA")
		}

		var mainProjects []Project
		err = json.Unmarshal([]byte(mainData), &mainProjects)
		if err != nil {
			fmt.Println("Error unmarshaling compareData as JSON:", err)
			os.Exit(1)
		}

		// compare the two objects at their current positions
		for i, original := range mainProjects {
			if i >= len(branchProjects) {
				break
			}
			updated := branchProjects[i]
			if !reflect.DeepEqual(original, updated) {
				validatedProjects = append(validatedProjects, updated)
			}
		}

		// check for any objects added to the array
		for i := len(mainProjects); i < len(branchProjects); i++ {
			added := branchProjects[i]
			validatedProjects = append(validatedProjects, added)
		}
	}

	var ids = map[string]bool{}

	validationFailed := false

	validate := validator.New()
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	validate.RegisterValidation("unique_id", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if ids[value] {
			return false
		}
		ids[value] = true
		return true
	})

	validate.RegisterValidation("alphanumeric_dashes", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		match, _ := regexp.MatchString("^[a-zA-Z0-9-]+$", value)
		return match
	})

	validate.RegisterValidation("no_emojis", func(fl validator.FieldLevel) bool {
		for i := 0; i < len(fl.Field().String()); {
			r, size := utf8.DecodeRuneInString(fl.Field().String()[i:])
			if r >= 0x1F600 && r <= 0x1F64F {
				return false
			}
			i += size
		}
		return true
	})

	validate.RegisterValidation("no_urls", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		re := regexp.MustCompile(`https?://\S+`)
		return !re.MatchString(value)
	})

	validate.RegisterValidation("no_redirect", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		res, err := client.Get(value)
		if err != nil {
			return false
		}
		defer res.Body.Close()
		return (res.StatusCode < 300 || res.StatusCode >= 400) ||
			// Medium instance articles do a 307
			res.StatusCode == 307
	})

	validate.RegisterValidation("date_for_article_video", func(fl validator.FieldLevel) bool {
		project := fl.Parent().Interface().(Project)
		if project.Type == "article" || project.Type == "video" {
			return validate.Var(project.Date, "datetime=2006-01-02") == nil
		}
		return true
	})

	fmt.Println("Validating projects:", validatedProjects)

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
