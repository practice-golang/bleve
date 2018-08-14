package main

import (
	"fmt"

	"github.com/blevesearch/bleve"
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

	index.Index("1", data)
	index.Index("2", data)

	// 검색 - 불완전한 단어는 검색되지 않으므로 정규표현식을 써야한다.
	// que := bleve.NewMatchQuery("lorem")
	que := bleve.NewRegexpQuery("(.*)하하(.*)")

	search := bleve.NewSearchRequest(que)
	searchResults, err := index.Search(search)
	if err != nil {
		panic(err)
	}

	fmt.Println(searchResults)

	for _, hit := range searchResults.Hits {
		fmt.Println(hit.ID)
		// data, err := index.GetInternal([]byte(hit.ID))
		data, err := index.Document(hit.ID)
		if err != nil {
			panic(err)
		}

		fmt.Println(&data.GoString())
	}
}
