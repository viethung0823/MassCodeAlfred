package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	aw "github.com/deanishe/awgo"
)

type Snippet struct {
	Name        string    `json:"name"`
	Id          string    `json:"id"`
	Description string    `json:"description"`
	IsDeleted   bool      `json:"isDeleted"`
	Folder      *Folder   `json:"folder"`
	Content     []Content `json:"content"`
	IsFavorites bool      `json:"isFavorites"`
	FolderId    string    `json:"folderId"`
	TagsIds     []string  `json:"tagsIds"`
	CreatedAt   int64     `json:"createdAt"`
	UpdatedAt   int64     `json:"updatedAt"`
}

type SnippetWithTags struct {
	Snippet
	TagValues []string `json:"tagValues"`
}

type Folder struct {
	Id              string  `json:"id"`
	Name            string  `json:"name"`
	DefaultLanguage string  `json:"defaultLanguage"`
	ParentId        *string `json:"parentId"`
	IsOpen          bool    `json:"isOpen"`
	IsSystem        bool    `json:"isSystem"`
	CreatedAt       int64   `json:"createdAt"`
	UpdatedAt       int64   `json:"updatedAt"`
	Index           int     `json:"index"`
	Icon            string  `json:"icon"`
}

type Tag struct {
	Name      string `json:"name"`
	Id        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type Content struct {
	Label    string `json:"label"`
	Language string `json:"language"`
	Value    string `json:"value"`
}

type Endpoints struct {
	GetSnippets string
	GetFolders  string
	GetTags     string
}

var APIEndpoints = Endpoints{
	GetSnippets: "http://localhost:3033/snippets/embed-folder",
	GetFolders:  "http://localhost:3033/folders",
	GetTags:     "http://localhost:3033/tags",
}

var wf = aw.New()
var query = ""

func main() {
	query = os.Args[1]
	searchMode := "Title"

	switch {
	case strings.HasPrefix(query, "f "):
		searchMode = "Folder"
	case strings.HasPrefix(query, "t "):
		searchMode = "Tag"
	}

	hanleSearch(searchMode)

	wf.SendFeedback()
}

func hanleSearch(searchMode string) {
	if searchMode != "Title" {
		queryParts := strings.SplitN(os.Args[1], " ", 2)
		if len(queryParts) > 1 {
			query = strings.TrimSpace(queryParts[1])
		} else {
			query = ""
		}
	}

	var snippetsData []Snippet
	err := FetchData(APIEndpoints.GetSnippets, &snippetsData)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	var tagsData []Tag
	err = FetchData(APIEndpoints.GetTags, &tagsData)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	snippetsWithTags := MergeTagsIntoSnippets(snippetsData, tagsData)

	for _, snippet := range snippetsWithTags {
		if !snippet.IsDeleted {
			title := snippet.Name
			subtitle := "Inbox"
			showFragmentLabel := len(snippet.Content) > 1
			if snippet.Folder != nil {
				subtitle = snippet.Folder.Name
			}

			urlScheme := "masscode://snippets/" + snippet.Id

			for _, fragment := range snippet.Content {
				item := wf.NewItem(title).
					Subtitle(subtitle).
					UID(snippet.Id).
					Var("description", snippet.Description).
					Var("snippet", fragment.Value).
					Arg(fragment.Value).
					Icon(&aw.Icon{Value: fmt.Sprintf(`icons/%s.svg`, (snippet.Folder.Icon))}).
					Valid(true)

				if showFragmentLabel {
					item.Subtitle(subtitle + " - " + fragment.Label)
				}

				if searchMode == "Folder" {
					item.Match(snippet.Folder.Name)
				}

				if searchMode == "Tag" {
					tagsString := strings.Join(snippet.TagValues, " , ")
					item.Match(tagsString)
				}

				item.Cmd().Arg(urlScheme).Subtitle("Open in MassCode")
				item.Opt().Subtitle("View snippet")
			}
		}
	}

	if query != "" {
		wf.Filter(query)
	}
}

// Utils
func FetchData(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode JSON: %v", err)
	}
	return nil
}

func MergeTagsIntoSnippets(snippetsData []Snippet, tagsData []Tag) []SnippetWithTags {
	// Create a map for quick lookup of tag names by their IDs.
	tagMap := make(map[string]string)
	for _, tag := range tagsData {
		tagMap[tag.Id] = tag.Name
	}

	// Merge tag data into snippetsData.
	var snippetsWithTags []SnippetWithTags
	for _, snippet := range snippetsData {
		var tagValues []string
		for _, tagId := range snippet.TagsIds {
			if tagName, exists := tagMap[tagId]; exists {
				tagValues = append(tagValues, tagName)
			}
		}
		snippetsWithTags = append(snippetsWithTags, SnippetWithTags{
			Snippet:   snippet,
			TagValues: tagValues,
		})
	}

	return snippetsWithTags
}
