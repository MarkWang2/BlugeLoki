package bluge_db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MarkWang2/BlugeLoki/storage/stores/util"
	"github.com/blugelabs/bluge"
	segment "github.com/blugelabs/bluge_segment_api"
	"log"
)

type BlugeDB struct {
	Name   string
	Folder string // snpseg
}

func NewDB(name string, path string) *BlugeDB {
	return &BlugeDB{Name: name, Folder: path} // "./snpseg"
}

type BlugeWriteBatch struct {
	Writes map[string]TableWrites
}

func NewWriteBatch() *BlugeWriteBatch {
	return &BlugeWriteBatch{
		Writes: map[string]TableWrites{},
	}
}

func (b *BlugeWriteBatch) getOrCreateTableWrites(tableName string) TableWrites {
	writes, ok := b.Writes[tableName]
	if !ok {
		writes = TableWrites{
			Puts: map[string]interface{}{},
		}
		b.Writes[tableName] = writes
	}

	return writes
}

func (b *BlugeWriteBatch) Delete(tableName, hashValue string, rangeValue []byte) {

}

func (b *BlugeWriteBatch) Add(tableName, hashValue string, rangeValue []byte, value []byte) {
	writes := b.getOrCreateTableWrites(tableName)

	key := hashValue
	writes.Puts[key] = string(value)
}

func (b *BlugeWriteBatch) AddJson(tableName string, data []byte) {
	writes := b.getOrCreateTableWrites(tableName)
	json.Unmarshal(data, &writes.Puts)
}

func (b *BlugeWriteBatch) AddEvent(tableName string, data util.MapStr) {
	writes := b.getOrCreateTableWrites(tableName)
	for k, v := range data {
		writes.Puts[k] = v
	}
}

type TableWrites struct {
	Puts util.MapStr // puts map[string][]byte
	//deletes map[string]struct{}
}

func (b *BlugeDB) WriteToDB(ctx context.Context, writes TableWrites) error {
	config := bluge.DefaultConfig(b.Folder + "/" + b.Name)
	writer, err := bluge.OpenWriter(config)
	if err != nil {
		log.Fatalf("error opening writer: %v", err)
	}
	defer writer.Close()

	doc := bluge.NewDocument("example") // can use server name

	for key, value := range writes.Puts {
		switch v := value.(type) {
		case nil:
			fmt.Println("x is nil") // here v has type interface{}
		case float64:
			fmt.Println("x is", v) // here v has type int
		case string:
			doc = doc.AddField(bluge.NewTextField(key, value.(string)).StoreValue())
		case bool:
			fmt.Println("x is bool or string") // here v has type interface{}
		default:
			fmt.Println("type unknown") // here v has type interface{}
		}
	}

	err = writer.Update(doc.ID(), doc)
	if err != nil {
		log.Fatalf("error updating document: %v", err)
	}
	return err
}

type IndexQuery struct {
	TableName string
	Query     bluge.Query
}

//visitor segment.segment.StoredFieldVisitor
//type segment.StoredFieldVisitor func(field string, value []byte) bool

// ctx context.Context, query chunk.IndexQuery, callback func(chunk.IndexQuery, chunk.ReadBatch) (shouldContinue bool)
func (b *BlugeDB) QueryDB(ctx context.Context, query IndexQuery, callback segment.StoredFieldVisitor) error {
	config := bluge.DefaultConfig(b.Folder + "/" + b.Name)
	writer, err := bluge.OpenWriter(config)
	reader, err := writer.Reader()
	if err != nil {
		log.Fatalf("error getting index reader: %v", err)
	}
	defer reader.Close()
	request := bluge.NewTopNSearch(10, query.Query).
		WithStandardAggregations()
	documentMatchIterator, err := reader.Search(context.Background(), request)
	if err != nil {
		log.Fatalf("error executing search: %v", err)
	}
	match, err := documentMatchIterator.Next()
	for err == nil && match != nil {
		err = match.VisitStoredFields(callback)
		if err != nil {
			log.Fatalf("error loading stored fields: %v", err)
		}
		match, err = documentMatchIterator.Next()
	}
	if err != nil {
		log.Fatalf("error iterator document matches: %v", err)
	}

	return nil
}

func (b *BlugeDB) Close() error {
	return nil
}

func (b *BlugeDB) Path() string {
	return b.Folder + "/" + b.Name
}
