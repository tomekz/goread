package rss

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
	"gopkg.in/yaml.v3"
)

// Rss will be used to structurize the rss feeds and categories
// it will usually be read from a file
type Rss struct {
	filePath   string     `yaml:"file_path"`
	Categories []Category `yaml:"categories"`
}

// Category will be used to structurize the rss feeds
type Category struct {
	Name          string
	Description   string `yaml:"desc"`
	Subscriptions []Feed `yaml:"subscriptions"`
}

// Feed is a single rss feed
type Feed struct {
	Name        string `yaml:"name"`
	Description string `yaml:"desc"`
	URL         string `yaml:"url"`
}

// ErrNotFound is returned when a feed or category is not found
var ErrNotFound = errors.New("not found")

// New will create a new Rss structure
func New(urlFilePath string) Rss {
	rss := Rss{filePath: urlFilePath}
	err := rss.loadFromFile()
	if err == nil {
		return rss
	}

	rss.Categories = append(rss.Categories, createBasicCategories()...)
	return rss
}

// loadFromFile will load the Rss structure from a file
func (rss *Rss) loadFromFile() error {
	// Check if the path is valid
	if rss.filePath == "" {
		// Get the default path
		path, err := getDefaultPath()
		if err != nil {
			return err
		}

		// Set the path
		rss.filePath = path
	}

	// Try to open the file
	fileContent, err := os.ReadFile(rss.filePath)
	if err != nil {
		return err
	}

	// Try to decode the file
	err = yaml.Unmarshal(fileContent, rss)
	if err != nil {
		return err
	}

	// Successfully loaded the file
	return nil
}

// Save will write the Rss structure to a file
func (rss *Rss) Save() error {
	fmt.Println("Saving rss file to", rss.filePath)
	for _, cat := range rss.Categories {
		fmt.Println("Category:", cat.Name)
	}

	// Try to marshall the data
	yamlData, err := yaml.Marshal(rss)
	if err != nil {
		return err
	}

	// Try to open the file, if it doesn't exist, create it
	file, err := os.OpenFile(rss.filePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		// Try to create the directory
		err = os.MkdirAll(filepath.Dir(rss.filePath), 0755)
		if err != nil {
			return err
		}

		// Try to create the file again
		file, err = os.Create(rss.filePath)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	// Write the data to the file
	_, err = file.Write(yamlData)
	if err != nil {
		return err
	}

	// Successfully wrote the file
	return nil
}

// GetCategories will return a alphabetically sorted list of all categories
func (rss Rss) GetCategories() []string {
	// Create a list of categories
	categories := make([]string, len(rss.Categories))
	for i, cat := range rss.Categories {
		categories[i] = cat.Name
	}

	// Sort the list
	sort.Strings(categories)

	// Return the list
	return categories
}

// GetFeeds will return a alphabetically sorted list of the feeds
// in a category denoted by the name
func (rss Rss) GetFeeds(categoryName string) ([]string, error) {
	// Find the category
	for _, cat := range rss.Categories {
		if cat.Name == categoryName {
			// Create a list of feeds
			feeds := make([]string, len(cat.Subscriptions))
			for i, feed := range cat.Subscriptions {
				feeds[i] = feed.Name
			}

			// Sort the list
			sort.Strings(feeds)

			// Return the list
			return feeds, nil
		}
	}

	// Category not found
	return nil, ErrNotFound
}

// GetFeedURL will return the url of a feed denoted by the name
func (rss Rss) GetFeedURL(feedName string) (string, error) {
	// Iterate over all categories
	for _, cat := range rss.Categories {
		// Iterate over all feeds
		for _, feed := range cat.Subscriptions {
			if feed.Name == feedName {
				return feed.URL, nil
			}
		}
	}

	// Feed not found
	return "", ErrNotFound
}

// Markdownize will return a string that can be used to display the rss feeds
func Markdownize(item gofeed.Item) string {
	var mdown string

	// Add the title
	mdown += "# " + item.Title + "\n "

	// If there are no authors, then don't add the author
	if item.Authors != nil {
		mdown += item.Authors[0].Name + "\n"
	}

	// Show when the article was published if available
	if item.PublishedParsed != nil {
		mdown += "\n"
		mdown += "Published: " + item.PublishedParsed.Format("2006-01-02 15:04:05")
	}

	// Convert the html to markdown
	mdown += "\n\n"
	mdown += htmlToMarkdown(item.Description)
	return mdown
}

// htmlToMarkdown converts html to markdown using the html-to-markdown library
func htmlToMarkdown(content string) string {
	converter := md.NewConverter("", true, nil)

	markdown, err := converter.ConvertString(content)
	if err != nil {
		panic(err)
	}

	return markdown
}

// HTMLToText converts html to text using the goquery library
func HTMLToText(content string) string {
	// Create a new document
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		panic(err)
	}

	// Return the text
	return doc.Text()
}

// getDefaultPath will return the default path for the urls file
func getDefaultPath() (string, error) {
	// Get the default config path
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	// Create the config path
	return filepath.Join(configDir, "goread", "urls.yml"), nil
}

// createBasicCategories will create some basic categories
func createBasicCategories() []Category {
	var categories []Category

	categories = append(categories, Category{
		Name:        "News",
		Description: "News from around the world",
	})

	categories = append(categories, Category{
		Name:        "Tech",
		Description: "Tech news",
	})

	categories[0].Subscriptions = append(categories[0].Subscriptions, Feed{
		Name:        "BBC",
		Description: "News from the BBC",
		URL:         "http://feeds.bbci.co.uk/news/rss.xml",
	})

	categories[1].Subscriptions = append(categories[1].Subscriptions, Feed{
		Name:        "Hacker News",
		Description: "News from Hacker News",
		URL:         "https://news.ycombinator.com/rss",
	})

	categories[1].Subscriptions = append(categories[1].Subscriptions, Feed{
		Name:        "Golang subreddit",
		Description: "News from the Golang subreddit",
		URL:         "https://www.reddit.com/r/golang/.rss",
	})

	return categories
}
