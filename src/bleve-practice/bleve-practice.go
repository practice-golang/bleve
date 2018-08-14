package main

import (
	"fmt"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
)

// Book - 도서 정보
type Book struct {
	ID     int
	Title  string
	Author string
}

func bleveInit(indexPath string) (bleve.Index, error) {
	index, err := bleve.Open(indexPath)
	if err != nil {
		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(indexPath, mapping)
		if err != nil {
			return nil, err
		}
	}

	return index, nil
}

func main() {
	index, err := bleveInit("storage")
	if err != nil {
		panic(err)
	}

	data := struct {
		Name string
	}{
		Name: "Lorem 크하하 Ipsum is simply dummy text of the printing and typesetting industry",
	}

	index.Index("First", data)
	index.Index("Second", data)

	// 검색
	// - 불완전한 단어는 검색되지 않으므로 정규표현식을 써야한다.
	// - 글자수 제한: 영어 4, 한글 2
	// que := bleve.NewMatchQuery("text")
	// que := bleve.NewRegexpQuery("(.*)text(.*)")
	que := bleve.NewRegexpQuery("(.*)하하(.*)")
	// que := bleve.NewRegexpQuery("(.*)")

	search := bleve.NewSearchRequest(que)
	searchResults, err := index.Search(search)
	if err != nil {
		panic(err)
	}

	if searchResults.Total == 0 {
		fmt.Println(searchResults)
	}

	for _, hit := range searchResults.Hits {
		doc, err := index.Document(hit.ID)
		if err != nil {
			panic(err)
		}
		// fmt.Println(doc.GoString())

		data := struct {
			ID     string                 `json:"id"`
			Fields map[string]interface{} `json:"fields"`
		}{
			ID:     hit.ID,
			Fields: map[string]interface{}{},
		}

		// Source: https://github.com/blevesearch/bleve/blob/master/http/doc_get.go
		for _, field := range doc.Fields {
			var newval interface{}

			switch field := field.(type) {
			case *document.TextField:
				newval = string(field.Value())
			case *document.NumericField:
				n, err := field.Number()
				if err == nil {
					newval = n
				}
			case *document.DateTimeField:
				d, err := field.DateTime()
				if err == nil {
					newval = d.Format(time.RFC3339Nano)
				}
			}

			existing, existed := data.Fields[field.Name()]
			if existed {
				switch existing := existing.(type) {
				case []interface{}:
					data.Fields[field.Name()] = append(existing, newval)
				case interface{}:
					arr := make([]interface{}, 2)
					arr[0] = existing
					arr[1] = newval
					data.Fields[field.Name()] = arr
				}
			} else {
				data.Fields[field.Name()] = newval
			}
		}

		// fmt.Println(data)
		fmt.Println(hit.ID, " : ", data.Fields["Name"])
	}
}
