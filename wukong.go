/*
	Author        : tuxpy
	Email         : q8886888@qq.com.com
	Create time   : 2018-01-12 09:51:00
	Filename      : wukong.go
	Description   :
*/

package main

import (
	"bufio"
	"dictcsv"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/huichen/wukong/engine"
	"github.com/huichen/wukong/types"
)

var (
	books_data = flag.String(
		"books_data",
		"./books.csv",
		"书本csv路径",
	)
	dictionaries = flag.String(
		"dictionaries",
		"./data/dictionary.txt",
		"分词字典文件",
	)
	stop_token_file = flag.String(
		"stop_token_file",
		"./data/stop_tokens.txt",
		"停用词文件",
	)
	searcher = engine.Engine{}
)

type BookScoringFields struct {
	Title string
	Score int
	ISBN  string
}

type BookScoringCriteria struct {
}

func (criteria BookScoringCriteria) Score(doc types.IndexedDocument, fields interface{}) []float32 {
	bf, ok := fields.(BookScoringFields)
	if !ok {
		return []float32{}
	}
	output := make([]float32, 1)
	output[0] = float32(bf.Score) / 5.0
	return output
}

func main() {
	flag.Parse()
	searcher.Init(types.EngineInitOptions{
		StopTokenFile:         *stop_token_file,
		SegmenterDictionaries: *dictionaries,
		//	UsePersistentStorage:    true,
		//	PersistentStorageFolder: "indexes_cache",
		//	IndexerBufferLength:     10,
		DefaultRankOptions: &types.RankOptions{
			MaxOutputs:      100,
			OutputOffset:    0,
			ScoringCriteria: BookScoringCriteria{},
		},
	})
	defer searcher.Close()

	reader := dictcsv.NewReaderFromFile(*books_data)
	records := reader.ReadAll()
	for i, record := range records {
		if len(record) == 0 {
			continue
		}
		score, _ := strconv.Atoi(record["comment_score"])

		searcher.IndexDocument(uint64(i), types.DocumentIndexData{
			Content: record["title"],
			Fields: BookScoringFields{
				Title: record["title"],
				Score: score,
				ISBN:  record["isbn"],
			},
			Labels: []string{record["isbn"], record["author"]},
		}, false)
	}
	searcher.FlushIndex()

	ioreader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		line, _, err := ioreader.ReadLine()
		if err != nil {
			break
		}
		output := searcher.Search(types.SearchRequest{
			Text: string(line),
		})
		var record map[string]string
		for _, doc := range output.Docs {
			record = records[doc.DocId]
			fmt.Println(doc.Scores, record["title"], record["comment_score"], record["source_url"])
		}

	}
}
