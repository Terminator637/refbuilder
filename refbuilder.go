package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/kljensen/snowball"
	"github.com/rakyll/statik/fs"
	_ "github.com/stiletto/refbuilder/statik"
	"html"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var spltr = regexp.MustCompile(`[ \s\t\.,(){}\[\]/*":]+`)
var language = "russian"
var stopWords = true

func processFile(fname string, defaultTitle string, depth int) (string, map[string]int, error) {
	fmt.Printf("Process file %q %d\n", fname, depth)
	var doc *goquery.Document
	f, err := os.Open(fname)
	if err != nil {
		doc, err = goquery.NewDocumentFromReader(strings.NewReader(""))
	} else {
		doc, err = goquery.NewDocumentFromReader(f)
		f.Close()
	}
	if err != nil {
		return "", nil, err
	}
	htmls := doc.Find("html")
	if htmls.Length() < 1 {
		doc.Wrap("html")
		htmls = doc.Find("html")
	}
	head := htmls.Find("head")
	if head.Length() < 1 {
		htmls.Contents().First().BeforeHtml("<head></head>")
		htmls = doc.Find("head")
	}
	titletext := doc.Find("title").Text()
	if titletext == "" {
		for _, h := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
			titletext = doc.Find(h).First().Text()
			if titletext != "" {
				break
			}
		}
		if titletext == "" {
			titletext = defaultTitle
		}
	}

	title := doc.Find("title")
	if title.Length() < 1 {
		head.AppendHtml("<title>dicks</title>")
		//fmt.Printf("WTF %s\n", ht2)
		title = doc.Find("title")
	}
	//html.EscapeString(titletext)
	title.ReplaceWithHtml("<title>" + html.EscapeString(titletext) + "</title>")

	script := head.Find(`script#ref_script`)
	scripthtml := fmt.Sprintf(`<script id="ref_script" src="%sassets/refbuilder.js"></script>`, strings.Repeat("../", depth))
	if script.Length() < 1 {
		head.Contents().First().BeforeHtml(scripthtml)
	} else {
		script.ReplaceWithHtml(scripthtml)
	}

	meta := head.Find(`meta[http-equiv="Content-Type"]`)
	metahtml := `<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>`
	if meta.Length() < 1 {
		head.Contents().First().BeforeHtml(metahtml)
	} else {
		meta.ReplaceWithHtml(metahtml)
	}

	ht, err := doc.Html()
	if err != nil {
		return "", nil, err
	}
	err = ioutil.WriteFile(fname+".tmp", []byte(ht), 0644)
	if err != nil {
		return "", nil, err
	}
	//os.Rename(fname,fname+".bak")
	err = os.Rename(fname+".tmp", fname)
	if err != nil {
		return "", nil, err
	}

	words := make(map[string]int)
	for _, w := range spltr.Split(doc.Find("body").Text(), -1) {
		if len(w) > 0 {
			w, err = snowball.Stem(w, language, stopWords)
			if err != nil {
				return "", nil, err
			}
			if len(w) > 0 {
				words[w] = words[w] + 1
			}
		}
	}

	return titletext, words, nil
}

type Tree []*TreeItem
type TreeItem struct {
	Id       string `json:"id"`
	Text     string `json:"text"`
	Children *Tree  `json:"children",omitempty`
	Type     string `json:"type"`
}

func processDir(base string, dir string, idx map[string]map[string]int, depth int) (*TreeItem, error) {
	fmt.Printf("Process dir %q %d\n", dir, depth)
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	files := make(map[string]*TreeItem)
	dirs := make(map[string]*TreeItem)
	for err == nil {
		var fis []os.FileInfo
		fis, err = f.Readdir(2)
		//fmt.Printf("rd %#v %#v\n", fis, err)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if fis != nil {
			for _, fi := range fis {
				finame := fi.Name()
				fipath := filepath.Join(dir, finame)
				firel, err := filepath.Rel(base, fipath)
				if err != nil {
					return nil, err
				}
				if fi.IsDir() {
					if finame != "assets" {
						subtree, err2 := processDir(base, fipath, idx, depth+1)
						if err2 != nil {
							return nil, err2
						}
						if subtree != nil {
							dirs[firel] = subtree
						}
					}
				} else {
					ext := strings.ToLower(filepath.Ext(finame))
					//fmt.Printf("Found %q %q\n", finame, ext)
					filower := strings.ToLower(finame)
					if (ext == ".html" || ext == ".htm") && filower != "index.html" && filower != "ref_frames.html" {
						title, words, err2 := processFile(fipath, finame, depth)
						if err2 != nil {
							return nil, err2
						}
						item := &TreeItem{
							Id:   "ref_node_" + firel,
							Text: title,
							Type: "file",
						}
						files[firel] = item
						for word, num := range words {
							widx, ok := idx[word]
							if !ok {
								widx = make(map[string]int)
								idx[word] = widx
							}
							widx[firel] = num
						}
					}
				}
			}
		}
	}
	if len(files) > 0 || len(dirs) > 0 {
		subtree := &TreeItem{Children: new(Tree), Type: "default"}
		for _, names := range []map[string]*TreeItem{dirs, files} {
			keys := make([]string, len(names))
			i := 0
			for k := range names {
				keys[i] = k
				i++
			}
			sort.Strings(keys)
			for _, k := range keys {
				*subtree.Children = append(*subtree.Children, names[k])
			}
		}
		title, _, err := processFile(filepath.Join(dir, "index.html"), filepath.Base(dir), depth)
		if err != nil {
			return nil, err
		}
		direl, err := filepath.Rel(base, dir)
		if err != nil {
			return nil, err
		}
		subtree.Id = "ref_node_" + direl
		subtree.Text = title
		return subtree, nil
	} else {
		return nil, nil
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <directory>\n", os.Args[0])
		os.Exit(1)
	}
	root := os.Args[1]
	words := make(map[string]map[string]int)
	tree, err := processDir(root, root, words, 0)
	if err == nil {
		tree.Type = "root"
		tree.Id = "ref_root"

		treeFile := filepath.Join(root, "tree.jsonp")
		treePrefix := "refbuilder.load_tree([\n"
		treeSuffix := "\n]);"

		data, err := json.Marshal(tree) //Indent(tree, "", "  ")
		if err != nil {
			panic(err.Error())
		}
		wrappedData := make([]byte, len(data)+len(treePrefix)+len(treeSuffix))
		copy(wrappedData, treePrefix)
		copy(wrappedData[len(treePrefix):], data)
		copy(wrappedData[len(wrappedData)-len(treeSuffix):], treeSuffix)
		err = ioutil.WriteFile(treeFile+".tmp", wrappedData, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to write tree data: %s", err)
			os.Exit(1)
		}
		err = os.Rename(treeFile+".tmp", treeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to replace tree data: %s", err)
			os.Exit(1)
		}

		idxFile := filepath.Join(root, "idx.jsonp")
		idxPrefix := "refbuilder.load_idx(\n"
		idxSuffix := "\n);"

		data, err = json.Marshal(words)
		if err != nil {
			panic(err.Error())
		}
		wrappedData = make([]byte, len(data)+len(idxPrefix)+len(idxSuffix))
		copy(wrappedData, idxPrefix)
		copy(wrappedData[len(idxPrefix):], data)
		copy(wrappedData[len(wrappedData)-len(idxSuffix):], idxSuffix)
		err = ioutil.WriteFile(idxFile+".tmp", wrappedData, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to write index data: %s", err)
			os.Exit(1)
		}
		err = os.Rename(idxFile+".tmp", idxFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to replace index data: %s", err)
			os.Exit(1)
		}

	} else {
		panic(err.Error())
	}
	err = os.MkdirAll(filepath.Join(root, "assets"), 0775)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create asset directory: %s\n", err)
		os.Exit(1)
	}
	sfs, _ := fs.New()
	index, err := sfs.Open("/index")
	if err != nil {
		panic(err.Error())
	}
	defer index.Close()
	scanner := bufio.NewScanner(index)
	for scanner.Scan() {
		finame := scanner.Text()
		asset := filepath.Join(root, "assets", finame)
		f, err := os.OpenFile(asset, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
		if err != nil {
			if os.IsExist(err) {
				fmt.Printf("Skipping %s as it already exists\n", asset)
				continue
			}
			fmt.Fprintf(os.Stderr, "Unable to write asset %s: %s\n", asset, err)
			os.Exit(1)
		}
		defer f.Close()
		src, err := sfs.Open("/" + finame)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to open asset %s even though it is in asset index\n", finame, err)
			os.Exit(2)
		}
		defer src.Close()
		_, err = io.Copy(f, src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to copy asset %s to %s: %s\n", finame, asset, err)
			os.Exit(2)
		}
		fmt.Printf("Asset %s -> %s\n", finame, asset)
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		panic(err.Error())
	}
}
