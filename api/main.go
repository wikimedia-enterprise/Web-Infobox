package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"github.com/wikimedia-enterprise/wme-sdk-go/pkg/api"
	"github.com/wikimedia-enterprise/wme-sdk-go/pkg/auth"
)

type InfoCard struct {
	Title       string            `json:"title"`
	URL         string            `json:"url,omitempty"`
	ImageURL    string            `json:"image_url,omitempty"`
	ImageAuthor string            `json:"image_author,omitempty"`
	Caption     string            `json:"caption,omitempty"`
	LastUpdated string            `json:"last_updated,omitempty"`
	Facts       map[string]string `json:"facts,omitempty"`
}

var apiClient api.API

func main() {
	godotenv.Load()
	ctx := context.Background()
	ath := auth.NewClient()
	lgn, err := ath.Login(ctx, &auth.LoginRequest{
		Username: os.Getenv("WIKI_USERNAME"),
		Password: os.Getenv("WIKI_PASSWORD"),
	})
	if err != nil {
		log.Fatalf("Failed to login to Wikimedia: %v", err)
	}

	apiClient = api.NewClient()
	apiClient.SetAccessToken(lgn.AccessToken)

	http.HandleFunc("/api/infobox", handleInfobox)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleInfobox(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	topic := r.URL.Query().Get("topic")
	if topic == "" {
		http.Error(w, `{"error": "Missing topic parameter"}`, http.StatusBadRequest)
		return
	}

	topic = strings.ReplaceAll(topic, " ", "_")

	req := &api.Request{
		Fields: []string{"name", "url", "article_body.html", "date_modified"},
		Filters: []*api.Filter{
			{Field: "is_part_of.identifier", Value: "enwiki"},
		},
	}

	scs, err := apiClient.GetArticles(context.Background(), topic, req)
	if err != nil || len(scs) == 0 {
		http.Error(w, `{"error": "Topic not found"}`, http.StatusNotFound)
		return
	}

	article := scs[0]
	card := InfoCard{
		Title: article.Name,
		URL:   article.URL,
		Facts: make(map[string]string),
	}

	if article.DateModified != nil {
		card.LastUpdated = article.DateModified.Format("January 2, 2006")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(article.ArticleBody.HTML))
	if err == nil {
		infobox := doc.Find("table.infobox").First()

		img := infobox.Find("img").First()
		if src, exists := img.Attr("src"); exists {
			if strings.HasPrefix(src, "//") {
				src = "https:" + src
			}
			card.ImageURL = src

			var filename string
			parentA := img.Closest("a")
			if href, ok := parentA.Attr("href"); ok && strings.Contains(href, "File:") {
				filename = href[strings.Index(href, "File:"):]
			} else if resource, ok := img.Attr("resource"); ok && strings.Contains(resource, "File:") {
				filename = resource[strings.Index(resource, "File:"):]
			}

			if filename != "" {
				filename = strings.Split(filename, "#")[0]
				filename = strings.Split(filename, "?")[0]

				if unescaped, err := url.PathUnescape(filename); err == nil {
					filename = unescaped
				}

				filename = strings.TrimPrefix(filename, "File:")
				filename = strings.TrimPrefix(filename, "file:")
				card.ImageAuthor = fetchImageAuthor("File:" + filename)
			}
		}

		card.Caption = strings.TrimSpace(infobox.Find(".infobox-caption").First().Text())

		var lastHeader string

		infobox.Find("tr").Each(func(i int, s *goquery.Selection) {
			if s.ParentsFiltered("table").Length() > 1 {
				return
			}

			s.Find("style, sup, .mw-empty-elt").Remove()
			s.Find("br").ReplaceWithHtml("\n")
			s.Find("li").Each(func(_ int, sel *goquery.Selection) {
				sel.PrependHtml("\n")
			})

			th := strings.TrimSpace(s.ChildrenFiltered("th").First().Text())
			td := strings.TrimSpace(s.ChildrenFiltered("td").First().Text())

			th = strings.ReplaceAll(th, "\u00a0", " ")
			td = strings.ReplaceAll(td, "\u00a0", " ")

			for strings.Contains(td, "\n\n") {
				td = strings.ReplaceAll(td, "\n\n", "\n")
			}

			if th != "" && s.ChildrenFiltered("td").Length() == 0 {
				lastHeader = th
				return
			}

			if th != "" && td != "" {
				card.Facts[th] = td
			} else if th == "" && td != "" && lastHeader != "" {
				card.Facts[lastHeader] = td
				lastHeader = ""
			}
		})

		var summaryText string
		var foundSummaryTable bool

		doc.Find("table").Each(func(_ int, table *goquery.Selection) {
			if foundSummaryTable {
				return
			}

			headers := table.Find("tr").First().Find("th, td")
			if headers.Length() >= 4 {
				h0 := strings.TrimSpace(headers.Eq(0).Text())
				h1 := strings.TrimSpace(headers.Eq(1).Text())

				if h0 == "Event" && (strings.Contains(h1, "1st") || strings.Contains(h1, "Gold")) {
					foundSummaryTable = true
					var stopParsing bool

					table.Find("tr").Each(func(rowIndex int, row *goquery.Selection) {
						if stopParsing || rowIndex == 0 {
							return
						}

						cells := row.Find("th, td")
						if cells.Length() > 0 && cells.Length() < 4 {
							stopParsing = true
							return
						}

						if cells.Length() >= 4 {
							event := strings.TrimSpace(cells.Eq(0).Text())
							if event != "Total" && event != "" && !strings.Contains(event, "Event") {
								gold := strings.TrimSpace(cells.Eq(1).Text())
								silver := strings.TrimSpace(cells.Eq(2).Text())
								bronze := strings.TrimSpace(cells.Eq(3).Text())

								summaryText += "\n" + event + ": " + gold + "\U0001F947 " + silver + "\U0001F948 " + bronze + "\U0001F949"
							}
						}
					})
				}
			}
		})

		if summaryText != "" {
			card.Facts["Medal record"] = strings.TrimSpace(summaryText)
		}
	}

	json.NewEncoder(w).Encode(card)
}

func fetchImageAuthor(filename string) string {
	titleParam := url.QueryEscape(filename)
	titleParam = strings.ReplaceAll(titleParam, "+", "%20")

	apiURL := "https://en.wikipedia.org/w/api.php?action=query&titles=" + titleParam + "&prop=imageinfo&iiprop=extmetadata&format=json"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return ""
	}

	req.Header.Set("User-Agent", "Web-Infobox-Scraper/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var result struct {
		Query struct {
			Pages map[string]struct {
				ImageInfo []struct {
					ExtMetadata map[string]struct {
						Value interface{} `json:"value"`
					} `json:"extmetadata"`
				} `json:"imageinfo"`
			} `json:"pages"`
		} `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}

	for _, page := range result.Query.Pages {
		if len(page.ImageInfo) > 0 {
			meta := page.ImageInfo[0].ExtMetadata
			var rawHTML string

			if artist, ok := meta["Artist"]; ok {
				if v, ok := artist.Value.(string); ok {
					rawHTML = v
				}
			} else if author, ok := meta["Author"]; ok {
				if v, ok := author.Value.(string); ok {
					rawHTML = v
				}
			} else if credit, ok := meta["Credit"]; ok {
				if v, ok := credit.Value.(string); ok {
					rawHTML = v
				}
			}

			if rawHTML != "" {
				doc, _ := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
				return strings.TrimSpace(doc.Text())
			}
		}
	}
	return ""
}
