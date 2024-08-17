package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	aw "github.com/deanishe/awgo"
)

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
}

type Content struct {
	Label    string `json:"label"`
	Language string `json:"language"`
	Value    string `json:"value"`
}

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

var wf = aw.New()

func main() {
	log.Print(os.Args)
	searchType := os.Args[1]
	query := os.Args[2]
	if searchType == "snippets" {
		searchSnippets()
	} else if searchType == "folders" {
		searchFolders()
	}

	if query != "" {
		wf.Filter(query)
	}

	wf.SendFeedback()
}

func searchSnippets() {
	resp, err := http.Get("http://localhost:3033/snippets/embed-folder")
	if err != nil {
		log.Fatalf("Failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	var data []Snippet
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatalf("Failed to decode JSON: %v", err)
	}

	for _, snippet := range data {
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
					Valid(true)

				if showFragmentLabel {
					item.Subtitle(subtitle + " - " + fragment.Label)
				}

				item.Cmd().Arg(urlScheme).Subtitle("Open in MassCode")
				item.Opt().Subtitle("View snippet")
			}
		}
	}
}

func searchFolders() {
	// wait for masscode update url scheme for folders
	resp, err := http.Get("http://localhost:3033/folders")
	if err != nil {
		log.Fatalf("Failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	var data []Folder
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatalf("Failed to decode JSON: %v", err)
	}

	for _, folder := range data {
		wf.NewItem(folder.Name).
			Subtitle(folder.DefaultLanguage).
			UID(folder.Id).
			Valid(true)
	}
}
