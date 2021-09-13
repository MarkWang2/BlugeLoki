package main

//
//
//import (
//	"archive/tar"
//	"bytes"
//	"compress/gzip"
//	"context"
//	"fmt"
//	"github.com/cortexproject/cortex/pkg/chunk/local"
//	shipper_util "github.com/grafana/loki/pkg/storage/stores/shipper/util"
//	"io"
//	"io/ioutil"
//	"log"
//	"os"
//	"path/filepath"
//	"strings"
//)
//import "github.com/blugelabs/bluge"
//
//func main() {
//	config := bluge.DefaultConfig("./snpseg")
//	writer, err := bluge.OpenWriter(config)
//	if err != nil {
//		log.Fatalf("error opening writer: %v", err)
//	}
//	defer writer.Close()
//
//	doc := bluge.NewDocument("example").
//		AddField(bluge.NewTextField("name", "bluge"))
//
//	err = writer.Update(doc.ID(), doc)
//	if err != nil {
//		log.Fatalf("error updating document: %v", err)
//	}
//
//	reader, err := writer.Reader()
//	if err != nil {
//		log.Fatalf("error getting index reader: %v", err)
//	}
//	defer reader.Close()
//
//	query := bluge.NewMatchQuery("bluge").SetField("name")
//	request := bluge.NewTopNSearch(10, query).
//		WithStandardAggregations()
//	documentMatchIterator, err := reader.Search(context.Background(), request)
//	if err != nil {
//		log.Fatalf("error executing search: %v", err)
//	}
//	match, err := documentMatchIterator.Next()
//	for err == nil && match != nil {
//		err = match.VisitStoredFields(func(field string, value []byte) bool {
//			if field == "_id" {
//				fmt.Printf("match: %s\n", string(value))
//			}
//			return true
//		})
//		if err != nil {
//			log.Fatalf("error loading stored fields: %v", err)
//		}
//		match, err = documentMatchIterator.Next()
//	}
//	if err != nil {
//		log.Fatalf("error iterator document matches: %v", err)
//	}
//
//	// tar + gzip
//	var buf bytes.Buffer
//	compress("./snpseg", &buf)
//
//	// write the .tar.gzip
//	fileToWrite, err := os.OpenFile("./compress.tar.gzip", os.O_CREATE|os.O_RDWR, os.FileMode(600))
//	if err != nil {
//		panic(err)
//	}
//	if _, err := io.Copy(fileToWrite, &buf); err != nil {
//		panic(err)
//	}
//
//	content, err := ioutil.ReadFile("obstore2")
//	red := bytes.NewReader(content)
//	// untar write
//	if err := decompress(red, "./uncompressHere/"); err != nil {
//		// probably delete uncompressHere?
//	}
//
//	content, err = ioutil.ReadFile("compress.tar.gzip")
//	fsObjectClient, err := local.NewFSObjectClient(local.FSConfig{Directory: "./obstore"})
//	fsObjectClient.PutObject(context.Background(), "path.Join(folder, fileName)2", bytes.NewReader(content))
//
//	pwd, _ := os.Getwd()
//	err = shipper_util.GetFileFromStorage(context.Background(), fsObjectClient, "path.Join(folder, fileName)2", pwd+"/obstore2")
//}
//
//func compress(src string, buf io.Writer) error {
//	// tar > gzip > buf
//	zr := gzip.NewWriter(buf)
//	tw := tar.NewWriter(zr)
//
//	// is file a folder?
//	fi, err := os.Stat(src)
//	if err != nil {
//		return err
//	}
//	mode := fi.Mode()
//	if mode.IsRegular() {
//		// get header
//		header, err := tar.FileInfoHeader(fi, src)
//		if err != nil {
//			return err
//		}
//		// write header
//		if err := tw.WriteHeader(header); err != nil {
//			return err
//		}
//		// get content
//		data, err := os.Open(src)
//		if err != nil {
//			return err
//		}
//		if _, err := io.Copy(tw, data); err != nil {
//			return err
//		}
//	} else if mode.IsDir() { // folder
//
//		// walk through every file in the folder
//		filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
//			// generate tar header
//			header, err := tar.FileInfoHeader(fi, file)
//			if err != nil {
//				return err
//			}
//
//			// must provide real name
//			// (see https://golang.org/src/archive/tar/common.go?#L626)
//			header.Name = filepath.ToSlash(file)
//
//			// write header
//			if err := tw.WriteHeader(header); err != nil {
//				return err
//			}
//			// if not a dir, write file content
//			if !fi.IsDir() {
//				data, err := os.Open(file)
//				if err != nil {
//					return err
//				}
//				if _, err := io.Copy(tw, data); err != nil {
//					return err
//				}
//			}
//			return nil
//		})
//	} else {
//		return fmt.Errorf("error: file type not supported")
//	}
//
//	// produce tar
//	if err := tw.Close(); err != nil {
//		return err
//	}
//	// produce gzip
//	if err := zr.Close(); err != nil {
//		return err
//	}
//	//
//	return nil
//}
//
//// check for path traversal and correct forward slashes
//func validRelPath(p string) bool {
//	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
//		return false
//	}
//	return true
//}
//
//func decompress(src io.Reader, dst string) error {
//	// ungzip
//	zr, err := gzip.NewReader(src)
//	if err != nil {
//		return err
//	}
//	// untar
//	tr := tar.NewReader(zr)
//
//	// uncompress each element
//	for {
//		header, err := tr.Next()
//		if err == io.EOF {
//			break // End of archive
//		}
//		if err != nil {
//			return err
//		}
//		target := header.Name
//
//		// validate name against path traversal
//		if !validRelPath(header.Name) {
//			return fmt.Errorf("tar contained invalid name error %q", target)
//		}
//
//		// add dst + re-format slashes according to system
//		target = filepath.Join(dst, header.Name)
//		// if no join is needed, replace with ToSlash:
//		// target = filepath.ToSlash(header.Name)
//
//		// check the type
//		switch header.Typeflag {
//
//		// if its a dir and it doesn't exist create it (with 0755 permission)
//		case tar.TypeDir:
//			if _, err := os.Stat(target); err != nil {
//				if err := os.MkdirAll(target, 0755); err != nil {
//					return err
//				}
//			}
//		// if it's a file create it (with same permission)
//		case tar.TypeReg:
//			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
//			if err != nil {
//				return err
//			}
//			// copy over contents
//			if _, err := io.Copy(fileToWrite, tr); err != nil {
//				return err
//			}
//			// manually close here after each file operation; defering would cause each file close
//			// to wait until all operations have completed.
//			fileToWrite.Close()
//		}
//	}
//
//	//
//	return nil
//}
//
////func decompress(src io.Reader, dst string) error {
////	// ungzip
////	zr, err := gzip.NewReader(src)
////	if err != nil {
////		return err
////	}
////	// untar
////	tr := tar.NewReader(zr)
////
////	// uncompress each element
////	for {
////		header, err := tr.Next()
////		if err == io.EOF {
////			break // End of archive
////		}
////		if err != nil {
////			return err
////		}
////		target := header.Name
////
////		// validate name against path traversal
////		if !validRelPath(header.Name) {
////			return fmt.Errorf("tar contained invalid name error %q", target)
////		}
////
////		// add dst + re-format slashes according to system
////		target = filepath.Join(dst, header.Name)
////		// if no join is needed, replace with ToSlash:
////		// target = filepath.ToSlash(header.Name)
////
////		// check the type
////		switch header.Typeflag {
////
////		// if its a dir and it doesn't exist create it (with 0755 permission)
////		case tar.TypeDir:
////			if _, err := os.Stat(target); err != nil {
////				if err := os.MkdirAll(target, 0755); err != nil {
////					return err
////				}
////			}
////		// if it's a file create it (with same permission)
////		case tar.TypeReg:
////			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
////			if err != nil {
////				return err
////			}
////			// copy over contents
////			if _, err := io.Copy(fileToWrite, tr); err != nil {
////				return err
////			}
////			// manually close here after each file operation; defering would cause each file close
////			// to wait until all operations have completed.
////			fileToWrite.Close()
////		}
////	}
////
////	//
////	return nil
////}
