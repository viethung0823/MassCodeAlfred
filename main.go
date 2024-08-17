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

type ExtendedSnippet struct {
	Snippet
	TagValues      []string `json:"tagValues"`
	FolderFullPath *string  `json:"parentFolderName,omitempty"`
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

	handleSearch(searchMode)

	wf.SendFeedback()
}

func handleSearch(searchMode string) {
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

	var foldersData []Folder
	err = FetchData(APIEndpoints.GetFolders, &foldersData)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	extenedSnippetDate := getExtendedSnippetDate(snippetsData, tagsData, foldersData)

	for _, snippet := range extenedSnippetDate {
		if !snippet.IsDeleted {
			title := snippet.Name
			subtitle := "Inbox"
			showFragmentLabel := len(snippet.Content) > 1
			if snippet.Folder != nil && snippet.FolderFullPath != nil {
				subtitle = *snippet.FolderFullPath
			}

			urlScheme := "masscode://snippets/" + snippet.Id

			for _, fragment := range snippet.Content {
				item := wf.NewItem(title).
					Subtitle(subtitle).
					UID(snippet.Id).
					Var("description", snippet.Description).
					Var("snippet", fragment.Value).
					Arg(fragment.Value).
					Valid(true)

				if showFragmentLabel {
					item.Subtitle(subtitle + " - " + fragment.Label)
				}

				if searchMode == "Folder" {
					item.Match(*snippet.FolderFullPath)
				}

				if searchMode == "Tag" {
					tagsString := strings.Join(snippet.TagValues, " , ")
					item.Match(tagsString)
				}

				iconPath := fmt.Sprintf("icons/%s.svg", snippet.Folder.Icon)
				if _, err := os.Stat(iconPath); !os.IsNotExist(err) {
					// Proceed with setting the icon
					item.Icon(&aw.Icon{Value: iconPath})
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

func getParentFolderPath(folderId string, folderMap map[string]Folder) string {
	folder, exists := folderMap[folderId]
	if !exists || folder.ParentId == nil {
		return folder.Name
	}
	parentPath := getParentFolderPath(*folder.ParentId, folderMap)
	return parentPath + "/" + folder.Name
}

func getExtendedSnippetDate(snippetsData []Snippet, tagsData []Tag, foldersData []Folder) []ExtendedSnippet {
	tagMap := make(map[string]string)
	for _, tag := range tagsData {
		tagMap[tag.Id] = tag.Name
	}

	folderMap := make(map[string]Folder)
	for _, folder := range foldersData {
		folderMap[folder.Id] = folder
	}

	var extendedSnippetData []ExtendedSnippet
	for _, snippet := range snippetsData {
		var tagValues []string
		for _, tagId := range snippet.TagsIds {
			if tagName, exists := tagMap[tagId]; exists {
				tagValues = append(tagValues, tagName)
			}
		}

		var folderFullPath *string
		if folder, exists := folderMap[snippet.FolderId]; exists {
			fullPath := getParentFolderPath(folder.Id, folderMap)
			folderFullPath = &fullPath
		}

		extendedSnippetData = append(extendedSnippetData, ExtendedSnippet{
			Snippet:        snippet,
			TagValues:      tagValues,
			FolderFullPath: folderFullPath,
		})
	}

	return extendedSnippetData
}
