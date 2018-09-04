package main

import (
	"fmt"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/document"
)

// Book - 도서 정보 / 사용 안 함.
// type Book struct {
// 	ID     int
// 	Title  string
// 	Author string
// }

// Data : 블레비 검색 결과 데이터
type Data struct {
	ID     string                 `json:"id"`
	Fields map[string]interface{} `json:"fields"`
}

// BleveInit : 초기화
func BleveInit(indexPath string) (bleve.Index, error) {
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

// GetResult : 결과 취득
func GetResult(searchResults *bleve.SearchResult, db bleve.Index) (result []*Data) {
	result = make([]*Data, 0)
	for _, hit := range searchResults.Hits {
		doc, err := db.Document(hit.ID)
		if err != nil {
			panic(err)
		}

		data := Data{
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
		// fmt.Println(data.ID, " : ", data.Fields["Title"], " / ", data.Fields["Author"], reflect.TypeOf(data.Fields["Author"]))
		result = append(result, &data)
	}
	return
}

func main() {
	// os.Setenv("GODEBUG", "gctrace=1")

	// Index 준비
	idx, err := BleveInit("storage")
	if err != nil {
		panic(err)
	}

	// data 준비
	// data := Book{
	// 	Title:  "Lorem 크하하 Ipsum is simply dummy text of the printing and typesetting industry",
	// 	Author: "WhoHaHaHeeHee",
	// }

	data := make(map[string]string)

	data["Title"] = "Lorem 크하하 Ipsum is simply dummy text of the printing and typesetting industry"
	data["Author"] = "WhoHaHaHeeHee"

	// 데이터 추가 ~= Crud
	idx.Index("First", data)
	idx.Index("Second", data)

	// 수정용 data 준비
	// dataMod := Book{
	// 	Title:  "Lorem 크하하 Ipsum is simply dummy text of the printing and typesetting industry BaBaBa",
	// 	Author: "WhoHa 으읨믜하하하",
	// }

	dataMod := make(map[string]string)

	dataMod["Title"] = "Lorem 크하하 Ipsum is simply dummy text of the printing and typesetting industry"
	dataMod["Author"] = "WhoHaHaHeeHee"

	// 데이터 수정 ~= crUd
	idx.Index("First", dataMod)

	// 다 쓴 map 정리
	data, dataMod = nil, nil

	// 검색
	// - 불완전한 단어는 검색되지 않으므로 정규표현식을 써야한다.
	// - 글자수 제한: 영어 4, 한글 2

	// Title, Author 전체 검색
	que1 := bleve.NewMatchQuery("simply")
	que2 := bleve.NewMatchQuery("text")
	que3 := bleve.NewRegexpQuery("(.*)하하(.*)")
	// que3 := bleve.NewRegexpQuery("(.*)") // 모든 목록 요청

	// Author 검색
	// 정규표현식은 소문자만 된다고 한다. - https://github.com/blevesearch/bleve/issues/989#issuecomment-415011983
	// que4 := bleve.NewMatchQuery("WhoHa")	// 이래 쓰면 안 된다.
	que4 := bleve.NewRegexpQuery("(.*)hoha(.*)") // 이래 쓰면 잘 된다.
	// que4 := bleve.NewRegexpQuery("(.*)읨믜(.*)")
	que4.SetField("Author")

	// 개별 생성한 쿼리 합치기
	que := bleve.NewConjunctionQuery()
	que.AddQuery(que1)
	que.AddQuery(que2)
	que.AddQuery(que3)
	// que.AddQuery(que4)

	// 검색 실행 ~= cRud
	search := bleve.NewSearchRequest(que)
	searchResults, err := idx.Search(search)
	if err != nil {
		panic(err)
	}

	// 이건 그냥 결과가 없을 때 표시
	if searchResults.Total == 0 {
		fmt.Println(searchResults)
	}

	// 결과 취득
	results := GetResult(searchResults, idx)

	// 결과 표시
	for _, r := range results {
		fmt.Println(r.ID, r.Fields["Title"], " / ", r.Fields["Author"])
	}

	// 인덱스 삭제
	idx.Delete("First")

	// 삭제 후 재검색 및 결과 표시
	fmt.Println("----")
	queAll := bleve.NewRegexpQuery("(.*)")
	que = bleve.NewConjunctionQuery()
	que.AddQuery(queAll)
	search = bleve.NewSearchRequest(que)
	searchResults, err = idx.Search(search)
	if err != nil {
		panic(err)
	}

	results = GetResult(searchResults, idx)
	for _, r := range results {
		fmt.Println(r.ID, r.Fields["Title"], " / ", r.Fields["Author"])
	}

	fmt.Println("끗")

	results = nil

	fmt.Println("리절츠 닐")
}
