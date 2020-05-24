package app

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

func getPostUrl(ctx *gin.Context) (string, error) {
	query := ctx.Request.URL.Query()
	if len(query["url"]) > 0 {
		return query["url"][0], nil
	} else {
		return "", errors.New("Empty url")
	}
}

/**
 * GitHub WikiのURLをパースして生のMarkdownが取得できるURLにする
 */
func ParseUrl(ctx *gin.Context) string {
	urlstr, nil := getPostUrl(ctx)
	u, err := url.Parse(urlstr)
	if err != nil {
		panic(err)
	}
	// この形式に変換する https://raw.github.com/wiki/user/repo/page.md?login=login&token=token
	path := ConvertWikiUrl(u.Path)
	rawUrl := "https://raw.github.com/wiki" + path + ".md"
	fmt.Println(rawUrl)
	return rawUrl
}

func ConvertWikiUrl(url string) string {
	// ホームの場合 https://github.com/yousan/toc-generator/wiki/
	// https://raw.githubusercontent.com/wiki/yousan/toc-generator/Home.md
	// それ以外の場合  https://github.com/yousan/toc-generator/wiki/testpage
	// https://raw.githubusercontent.com/wiki/yousan/toc-generator/testpage.md
	r := regexp.MustCompile(`(.*)wiki(/[^/]+)?/?$`)
	regexedWikiPagename := r.ReplaceAllString(url, "$2")
	var wikiPagename string
	if len(regexedWikiPagename) == 0 {
		wikiPagename = "Home"
	} else {
		wikiPagename = regexedWikiPagename
	}
	str := r.ReplaceAllString(url, "$1")
	fmt.Println(str)
	return str + wikiPagename
}

/*
URLデータを読み込む
*/
func getContent(url string) string {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	byteArray, _ := ioutil.ReadAll(resp.Body)
	return string(byteArray)
}

/**

# a
## b
### c
## d
# e

* a
* * b

 */
func ParseMarkdownToUl(content string) []string {
	var ret []string

	var s = strings.Split(content, "\n")
	for i := 0; i < len(s); i ++ {
		r := regexp.MustCompile(`^(#+)(.*)$`)
		if r.Match([]byte(s[i])) { // マッチした場合のみ反応させる
			headingMark := r.ReplaceAllString(s[i], "$1")
			headingStr := r.ReplaceAllString(s[i], "$2")
			if len(headingMark) > 0 {
				fmt.Printf("%s %d\n", headingMark, len(headingMark))
				ret = append(ret, ToUL(len(headingMark), headingStr))
			}
		}
	}
	return ret
}

func ToUL(num int, heading string) string {
	var ret string
	for i := 0; i < num - 1; i++ {
		ret = ret + "  "
	}
	ret = ret + "* " + heading
	return ret
}


func dirwalk(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}

	return paths
}

func Default() *gin.Engine {
	router := gin.Default()

	// osを使う方法
	// path, _ := os.Getwd()
	// router.LoadHTMLGlob(path + "/templates/*.html")

	// runtimeを使う方法
	//pc, file, line, _  := runtime.Caller(1)
	//fmt.Printf("Called from %s, line #%d, func: %v\n",
	//	file, line, runtime.FuncForPC(pc).Name())
	//d, _ := filepath.Split(file)
	//fmt.Print(d)
	//// fmt.Println(dirwalk(d))
	//router.LoadHTMLGlob(d + "templates/*.html")

	// loadHTMLFilesバージョン
	// 相対パス
	// router.LoadHTMLFiles("templates/index.html")

	// 絶対パス1
	// path, _ := os.Getwd()
	// router.LoadHTMLFiles(path + "/templates/index.html")
	// 結果 open /var/task/templates/index.html: no such file or directory

	// 絶対パス2
	_, file, _, _  := runtime.Caller(1)
	//fmt.Printf("Called from %s, line #%d, func: %v\n",
	//	file, line, runtime.FuncForPC(pc).Name())
	d, _ := filepath.Split(file)
	//fmt.Print(d)
	//// fmt.Println(dirwalk(d))
	router.LoadHTMLFiles(d + "templates/index.html")


	data := "Hello Go/Gin!!"

	router.GET("/", func(ctx *gin.Context) {
		url, _ := getPostUrl(ctx)
		vars := make(map[string]string)
		vars["url"] = url
		if len(url) > 0 {
			rawurl := ParseUrl(ctx)
			vars["rawurl"] = rawurl
			content := getContent(rawurl)
			vars["rawbody"] = content
			uls := ParseMarkdownToUl(content)
			toc := "# ToC\n"
			for i := 0; i<len(uls); i++ {
				toc = toc + uls[i] + "\n"
			}
			vars["toc"] = toc
		}
		ctx.HTML(200,"index.html",
			vars)
	})

	router.GET("/url", func(ctx *gin.Context) {
		rawurl := ParseUrl(ctx)
		content := getContent(rawurl)
		ParseMarkdownToUl(content)
		ctx.HTML(200, "index.html", gin.H{"data": data})
	})

	// 開発用の出力分
	lines, _ := readBytes("testpage.md")
	uls := ParseMarkdownToUl(lines)
	fmt.Printf("%s\n", uls)

//	router.Run()
	return router
}

func readBytes(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var lines string

	b := make([]byte, 10)
	for {
		c, err := file.Read(b)
		if c == 0 {
			break
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		// line := string(b[:c])
		lines = lines + string(b[:c])
		// fmt.Print(line)
	}
	return lines, nil
}