// Package blog is a simple blog based on Gorilla mux and blackfriday markdown.
package blog

import (
    "os"
    "fmt"
    "log"
    "time"
    "strings"
    "net/http"
    "html/template"
    "path/filepath"
    "io/ioutil"

    "github.com/gorilla/mux"
    "github.com/russross/blackfriday"
)

const (
    // dateFormat is the expected format of post filename dates.
    dateFormat = "2006-01-02"
    // fieldSeparator is what is used to split post filenames.
    fieldSeparator = "."
)

// Blog is a blog site Handler..
type Blog struct {
    // Router is the mux router which Blog resides under.
    Router *mux.Router
    // TplPath is the relative path to the HTML template for the Blog.
    TplPath string
    // PostDir is the relative path to the directory containing blog posts.
    PostDir string
    // PageSize is the number of posts to a page.
    PageSize int
}

// NewBlog returns a new Blog instance.
func NewBlog(router *mux.Router, tplPath, postDir string,
                pageSize int) (blog *Blog) {
    blog = &Blog{Router: router, TplPath: tplPath,
                    PostDir: postDir, PageSize: pageSize}
    // Hook up paths for the main blog and post permalinks.
    router.Handle("/", blog)
    router.Handle("/post/{year:[0-9]+}/{month:[0-9]+}/{day:[0-9]+}/{name}/",
                    blog).Name("post")
    return
}

func (self Blog) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    tpl, err := template.ParseFiles(self.TplPath)
    if err != nil {
        log.Fatal(err)
    }

    // data is the template data for the Blog.
    data := struct {
        // Posts is the slice of posts for this blog page.
        Posts []post
    } {}

    // Pull out {…} variables from muxer.
    vars := mux.Vars(r)

    switch mux.CurrentRoute(r).GetName() {
    case "post":
        data.Posts = make([]post, 1)
        newPost, err := self.getPost(vars["year"], vars["month"],
                                    vars["day"], vars["name"])
        if err != nil {
            log.Fatal(err)
        }
        data.Posts[0] = *newPost
    default:
        data.Posts, err = self.getPage(0)
    }

    tpl.Execute(w, data)
}

// post is a single blog entry.
type post struct {
    // Name is the name/title of the post.
    Name string
    // Body is the content of the post.
    Body template.HTML
    // Date is the date the post was published.
    Date time.Time
    // Permalink is a permanent URL pointing to this post.
    Permalink string
}

// newPost returns a new post instance.
func (self Blog) newPost(path string, info os.FileInfo) (*post, error) {

    // Extract the fields from the filename, assuming a format of:
    //  YYYY-MM-DD.NAME.md
    // which produces a slice of:
    //  [0] => YYYY-MM-DD
    //  [1] => NAME
    //  [2] => .md
    // where len(fields) >= 2
    nameFields := strings.Split(info.Name(), fieldSeparator)
    var dateField, nameField string
    if len(nameFields) >= 2 {
        dateField = nameFields[0]
        nameField = nameFields[1]
    }
    postDate, err := time.ParseInLocation(dateFormat, dateField, time.UTC)
    if err != nil {
        return nil, err
    }
    postName := nameField

    // Create a permalink to this post using the named route "post".
    postPermalinkURL, err := self.Router.Get("post").
        URL("year", fmt.Sprintf("%d", postDate.Year()),
            "month", fmt.Sprintf("%d", postDate.Month()),
            "day", fmt.Sprintf("%d", postDate.Day()),
            "name", postName)
    if err != nil {
        return nil, err
    }
    postPermalink := postPermalinkURL.String()

    // Get file, read contents, render as HTML.
    postFile, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    postMarkdown, err := ioutil.ReadAll(postFile)
    if err != nil {
        return nil, err
    }
    postHTML := template.HTML(blackfriday.MarkdownCommon(postMarkdown))

    return &post{Name: postName, Body: postHTML, Date: postDate,
                    Permalink: postPermalink}, nil
}

// getPost retrieves a post from the given identifying information.
func (self Blog) getPost(year, month, day, name string) (*post, error) {
    // Use permalink info to construct file path.
    path := self.PostDir + fmt.Sprintf("%04s-%02s-%02s.%s.md", year, month, day, name);
    // Make sure the file exists and get info.
    info, err := os.Stat(path)
    if err != nil {
        return nil, err
    }
    newPost, err := self.newPost(path, info)
    if err != nil {
        return nil, err
    }
    return newPost, nil
}

// getPage retrieves one page of posts.
// XXX: pagination not actually implemented yet; page parameter ignored.
func (self Blog) getPage(page int) ([]post, error) {
    var posts []post
    // A function to walk the post directory and put together our slice.
    buildPosts := func(path string, info os.FileInfo, err error) error {
        if err != nil {
            // Don't die, just note problem and move on.
            log.Printf("Error walking posts: %v\n", err)
            return nil
        }

        if !info.IsDir() {
            newPost, err := self.newPost(path, info)
            if err != nil {
                return err
            }
            posts = append(posts, *newPost)
        }

        return nil
    }
    err := filepath.Walk(self.PostDir, buildPosts)
    if err != nil {
        return nil, err
    }

    // Reverse post order so most recent is shown first.
    for i, j := 0, len(posts) - 1; i < j; i, j = i + 1, j - 1 {
        posts[i], posts[j] = posts[j], posts[i]
    }

    return posts, nil
}

