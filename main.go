package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

// ApplicationConfig configuration
type configuratie struct {
	Baseurl   string
	Jsonfiles string
}

type datasetsStruct struct {
	Datasets []string
}

var configPath string
var config configuratie = *getConfig()
var combinedJson string

func main() {
	time.Sleep(5 * time.Second)
	json := getCombinedJson()
	if "" != json {
		combinedJson = json
	}
	go watch(config.Jsonfiles, &combinedJson)

	router := mux.NewRouter()
	router.HandleFunc("/", healthCeck)
	router.HandleFunc("/themes", listThemes)

	fs := http.FileServer(http.Dir("./swaggerui"))
	router.PathPrefix("/swaggerui/").Handler(http.StripPrefix("/swaggerui/", fs))

	if err := http.ListenAndServe(":80", router); err != nil {
		panic(err)
	}
}

func getConfig() *configuratie {
	cliConfigPath := flag.String("c", "none", "path of the configuration file")
	flag.Parse()

	if *cliConfigPath == "none" {
		log.Error("No configuration file...")
	} else {
		configPath = *cliConfigPath
	}

	fmt.Println("Starting viewer-config-service service...")

	configFile, err := ioutil.ReadFile(configPath)

	if err != nil {
		log.Fatalf("Could not read config")
	}

	var config = &configuratie{}
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		log.Fatal("Unmarshal: %v", err)
	}
	fmt.Println(config)
	return config
}

func isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func getThemePaths() []os.FileInfo {
	files, err := ioutil.ReadDir(config.Jsonfiles)
	if err != nil {
		panic(err)
	}
	dirs := []os.FileInfo{}
	for _, f := range files {
		if f.IsDir() {
			dirs = append(dirs, f)
		}
	}
	return dirs
}

func getBaseDatasetPaths(themeDir string) os.FileInfo {
	fileInfo, err := os.Lstat(config.Jsonfiles + themeDir + "/base.json")
	if err != nil {
		panic(err)
	}
	return fileInfo
}

func getDatasetPaths(themeDir string) []os.FileInfo {
	paths, err := ioutil.ReadDir(config.Jsonfiles + themeDir)
	if err != nil {
		panic(err)
	}
	files := []os.FileInfo{}
	for _, f := range paths {
		if !f.IsDir() {
			if err != nil {
				panic(err)
			}
			files = append(files, f)
		}
	}
	return files
}

func healthCeck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed: "+r.Method, 405)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"health\" : \"OK\"}"))
}

func listThemes(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "method not allowed: "+r.Method, 405)
	}
	w.Header().Set("Content-Type", "application/json")
	// tmp := <- combinedJson
	w.Write([]byte(combinedJson))
}

func getCombinedJson() string {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	log.Info("Reloading json files")
	var themes []string
	themeDirs := getThemePaths()
	for _, t := range themeDirs {

		// base.json's uitlezen
		base := getBaseDatasetPaths(t.Name())
		buffer := make([]byte, base.Size())
		var f interface{}
		json.Unmarshal(buffer, &f)

		// Overige json files uit bijbehorende map inlezen
		var datasets []string
		datasetPaths := getDatasetPaths(t.Name())
		for _, d := range datasetPaths {
			if !strings.Contains(d.Name(), "base.json") {
				path := config.Jsonfiles + t.Name() + "/" + d.Name()
				stat, err := os.Stat(path)
				if err != nil {
					log.Error(err)
					continue
				}
				if stat.IsDir() {
					continue
				}
				buffer, err := ioutil.ReadFile(path)
				if err != nil {
					log.Error(err)
					continue
				}
				json := string(buffer)

				if !isJSON(json) {
					log.Error("INVALID JSON: " + json)
					continue
				}
				datasets = append(datasets, parseJson(d, t))
			}
		}
		// base.json's vullen
		var theme bytes.Buffer
		tmpl := template.Must(template.New("base.json").Funcs(template.FuncMap{"StringsJoin": strings.Join}).ParseFiles(config.Jsonfiles + t.Name() + "/" + base.Name()))
		err := tmpl.Execute(&theme, datasetsStruct{Datasets: datasets})
		if err != nil {
			log.Error(err)
			continue
		}
		// base.json's samenvoegen in een lijst
		// Alle \r\n verwijderen uit de string
		var strippedString = strings.Replace(theme.String(), "\r\n", "", -1)
		themes = append(themes, strippedString)
	}
	return "[" + strings.Join(themes, ",") + "]"
}

func parseJson(d os.FileInfo, t os.FileInfo) string {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	var json bytes.Buffer
	tmpl := template.Must(template.New(d.Name()).Funcs(template.FuncMap{"StringsJoin": strings.Join}).ParseFiles(config.Jsonfiles + t.Name() + "/" + d.Name()))
	err := tmpl.Execute(&json, config)
	if err != nil {
		log.Error(err)
	}
	// base.json's samenvoegen in een lijst
	// Alle \r\n verwijderen uit de string
	return strings.Replace(json.String(), "\r\n", "", -1)
}
