package docs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Docs struct {
	token string
}

type ItemID struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
}

type CollectionsResponse struct {
	Collections struct {
		Page  int      `json:"page"`
		Pages int      `json:"pages"`
		Items []ItemID `json:"items"`
	} `json:"collections"`
}

type ArticlesResponse struct {
	Articles struct {
		Page  int      `json:"page"`
		Pages int      `json:"pages"`
		Items []ItemID `json:"items"`
	} `json:"articles"`
}

type Article struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Text      string `json:"text"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updatedAt"`
}

type ArticleResponse struct {
	Article Article `json:"article"`
}

func NewDocs(token string) *Docs {
	return &Docs{token}
}

func (d *Docs) MakeRequest(method string, endpoint string) *http.Response {
	client := &http.Client{}

	req, err := http.NewRequest(method, "https://docsapi.helpscout.net/v1"+endpoint, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil
	}

	req.SetBasicAuth(d.token, "X")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil
	}

	return resp
}

func (d *Docs) CollectionIDS(page int) []string {
	resp := d.MakeRequest("GET", "/collections?page="+strconv.Itoa(page))
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	var collectionsResponse CollectionsResponse
	if err := json.Unmarshal(body, &collectionsResponse); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return nil
	}

	var collectionIDs []string
	for _, collection := range collectionsResponse.Collections.Items {
		collectionIDs = append(collectionIDs, collection.ID)
	}

	if page < collectionsResponse.Collections.Pages {
		results := d.CollectionIDS(page + 1)
		collectionIDs = append(collectionIDs, results...)
	}

	return collectionIDs
}

func (d *Docs) ArticleNumbers(collectionID string, page int) []int {
	resp := d.MakeRequest("GET", "/collections/"+collectionID+"/articles?page="+strconv.Itoa(page))
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	var articlesResponse ArticlesResponse
	if err := json.Unmarshal(body, &articlesResponse); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return nil
	}

	var articleNumbers []int
	for _, article := range articlesResponse.Articles.Items {
		articleNumbers = append(articleNumbers, article.Number)
	}

	if page < articlesResponse.Articles.Pages {
		results := d.ArticleNumbers(collectionID, page + 1)
		articleNumbers = append(articleNumbers, results...)
	}

	return articleNumbers
}

func (d *Docs) Article(number int) *Article {
	resp := d.MakeRequest("GET", "/articles/"+strconv.Itoa(number))
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}

	var articleResponse ArticleResponse
	if err := json.Unmarshal(body, &articleResponse); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return nil
	}

	return &articleResponse.Article
}
