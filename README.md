# blog #

A simple blog written in [Go](http://golang.org/) and based on [Gorilla
mux](https://github.com/gorilla/mux) and
[blackfriday](https://github.com/russross/blackfriday).

## Rationale ##

There are plenty of blog frameworks out there, both static and dynamic, which
offer a lot of great functionality and robustness out of the box. This is not
necessarily one of those frameworks.

This package is for those who want a simple blog to drop into an existing 
Go-based website. Adding a page is as simple as writting a Golang HTML template,
dropping semantically-named blog posts written in markdown into a directory,
and adding a new route to your mux.

## Usage ##

Installation:
```
$ go get github.com/scott-linder/blog
```

In your template:
```HTML
{{range $post := .Posts}}
<article>
<a href="{{$post.Permalink}}><h1>{{$post.Name}}</h1></a>
{{$post.Body}}
</article>
{{else}}
<p>No posts.</p>
{{end}}
```

In your source:
```Go
r := mux.NewRouter()
sb := r.PathPrefix("/blog").Subrouter()
blog.NewBlog("name", sb, "path/to/template.tpl", "path/to/posts/", pageSize)
```

