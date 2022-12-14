package cyoa

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

func init() {
	tpl = template.Must(template.ParseFiles("index.html"))
}

var tpl *template.Template

// HandlerOption are used with the NewHandler function to
// configure the http.Handler returned.
type HandlerOption func(h *handler)

// WithTemplate is an option to provide a custom template to
// be used when rendering stories.
func WithTemplate(t *template.Template) HandlerOption {
	return func(h *handler) {
		h.t = t
	}
}

// WithPathFunc is an option to provide a custom function
// for processing the story chapter from the incoming request.
// This should probably be named something like "ChapterParser"
// but leaving it as is since this what I used in the video.
func WithPathFunc(fn func(r *http.Request) string) HandlerOption {
	return func(h *handler) {
		h.pathFn = fn
	}
}

// NewHandler will construct an http.Handler that will render
// the story provided.
// The default handler will use the full path (minus the / prefix)
// as the chapter name, defaulting to "intro" if the path is
// empty. The default template creates option links that follow
// this pattern.
func NewHandler(s Story, opts ...HandlerOption) http.Handler {
	h := handler{s, tpl, defaultPathFn}
	for _, opt := range opts {
		opt(&h)
	}
	return h
}

type handler struct {
	s      Story
	t      *template.Template
	pathFn func(r *http.Request) string
}

func defaultPathFn(r *http.Request) string {
	path := strings.TrimSpace(r.URL.Path)
	if path == "" || path == "/" {
		path = "/intro"
	}
	return path[1:]
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := h.pathFn(r)

	if chapter, ok := h.s[path]; ok {
		err := h.t.Execute(w, chapter)
		if err != nil {
			log.Printf("%v", err)
			http.Error(w, "Something went wrong...", http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "Chapter not found.", http.StatusNotFound)
}

// JSONStory will decode a story using the incoming reader
// and the encoding/json package. It is assumed that the
// provided reader has the story stored in JSON.
func JSONStory(r io.Reader) (Story, error) {
	d := json.NewDecoder(r)
	var story Story
	if err := d.Decode(&story); err != nil {
		return nil, err
	}
	return story, nil
}

// Story represents a Choose Your Own Adventure story.
// Each key is the name of a story chapter (aka "arc"), and
// each value is a Chapter.
type Story map[string]Chapter

// Chapter represents a CYOA story chapter (or "arc"). Each
// chapter includes its title, the paragraphs it is composed
// of, and options available for the reader to take at the
// end of the chapter. If the options are empty it is
// assumed that you have reached the end of that particular
// story path.
type Chapter struct {
	Title      string   `json:"title"`
	Paragraphs []string `json:"story"`
	Options    []Option `json:"options"`
}

// Option represents a choice offered at the end of a story
// chapter. Text is the visible text end users will see,
// while the Chapter field will be the key to a chapter
// stored in the Story object this chapter was found in.
type Option struct {
	Text    string `json:"text"`
	Chapter string `json:"arc"`
}
