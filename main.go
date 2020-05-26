package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/unidoc/unioffice/common"
	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/measurement"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var reportLink string

// Init config program
func (conf *ConfigFile) init(path string) {

	fileConfig, err := os.Open(path)
	if err != nil {
		log.Println(err)
	}
	defer fileConfig.Close()
	decoder := json.NewDecoder(fileConfig)
	err = decoder.Decode(&conf)
	if err != nil {
		log.Println(err)
	}

	// config time zone for shift time
	conf.TimeZone[0] = "Europe/Dublin"
	conf.TimeZone[1] = "Europe/Amsterdam"
	conf.TimeZone[2] = "Asia/Amman"
	conf.TimeZone[3] = "Europe/Moscow"
	conf.TimeZone[4] = "Asia/Tbilisi"
}

//TODO default
func getOrDefault(input interface{}, config interface{}) interface{} {
	//fmt.Printf("value=%v type=%t\n", input, input)
	switch input.(type) {
	case string:
		if input.(string) == "" {
			if config.(string) != "" {
				return config
			}
		}
	case int:
		if input.(int) == 0 {
			if config.(int) != 0 {
				return config
			}
		}
	default:
		if input == nil {
			if config != nil {
				return config
			}
		}
	}
	return input
}

// get info from grafana
func getInfo(config ConfigFile, web ConfigWeb) {
	var filePath string
	var strUrl string

	// config time zone for shift time
	TimeZoneGrafana := make(map[string]int64)
	TimeZoneGrafana["Europe/Dublin"] = 0       // UTC
	TimeZoneGrafana["Europe/Amsterdam"] = 3600 // UTC + 1
	TimeZoneGrafana["Asia/Amman"] = 7200       // UTC + 2
	TimeZoneGrafana["Europe/Moscow"] = 10800   // shift time, Asia/Riyadh = UTC + 3 = 10800 timestamp
	TimeZoneGrafana["Asia/Tbilisi"] = 14400    //UTC + 4

	// convert time to timestamp
	web.TimeFrom = convertTime(web.TimeFrom, TimeZoneGrafana[web.TimeZone])
	web.TimeTo = convertTime(web.TimeTo, TimeZoneGrafana[web.TimeZone])

	// create cookie
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}
	// сохроняем все картинки из графаны в файлы/папки на сервере
	// идём по всем проектам в конфиге
	for _, project := range config.Projects {

		// проверка названия проекта
		if project.Name != web.Project {
			continue
		}

		// Load default values from configs field
		configs := project.Configs

		// Handle every instruction
		for _, instr := range project.Instructions {
			// Check is there there instructions for grafana, then download images for every panelId
			dash := instr.Grafana
			if dash.PanelsID != nil {
				// разрезаем строку адресан на блоки для дальнейшей работы
				// на входе вот такая строка http://127.0.0.1:3000/d/Gm189NKmz/new-dashboard-copy
				host := getOrDefault(dash.Hostname, configs.Grafana.Hostname).(string)
				host = regexp.MustCompile(`(^.+//.+?)(/|$)`).FindStringSubmatch(host)[1]
				reg := regexp.MustCompile(`(.*/|^)(.+?/.+?)(\?|$)`)
				dashboard := getOrDefault(dash.Dashboard, configs.Grafana.Dashboard).(string)
				dPath := reg.FindStringSubmatch(dashboard)
				dashName := strings.Split(dPath[2], "/")[1]

				// создаем папку для графиков
				filePath = fmt.Sprintf("imgs/%s/%s", project.Name, dashName)
				createFolder(filePath)

				for _, panelId := range dash.PanelsID {
					// http://127.0.0.1:3000/render/d-solo/Gm189NKmz/new-dashboard-copy?panelId=2&orgId=1&from=1532423409083&to=1532431126864&width=1000&height=500
					strUrl = fmt.Sprintf("%s/render/d-solo/%s?panelId=%s&from=%s&to=%s&tz=%s", host, dPath[2], panelId, web.TimeFrom, web.TimeTo, web.TimeZone)

					Width := getOrDefault(dash.Width, configs.Grafana.Width).(int)
					Height := getOrDefault(dash.Height, configs.Grafana.Height).(int)
					if (Width != 0) && (Height != 0) {
						strUrl += "&width=" + strconv.Itoa(Width) + "&height=" + strconv.Itoa(Height)
					} else {
						// default size image
						strUrl += "&width=1100&height=500"
					}

					log.Printf("Downloading %s", strUrl) // Verify what url is used for request

					req, err := http.NewRequest("GET", strUrl, nil)
					if err != nil {
						log.Printf("Error open welcome web, %s", err)
						continue
					}
					req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", project.Key))
					resp, err := client.Do(req)
					if err != nil {
						log.Printf("Error open web, %s", err)
						continue
					}

					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Printf("Error read file, %s", err)
						continue
					}
					resp.Body.Close()

					// create image
					fileHandle, _ := os.Create(filePath + "/" + panelId + ".png")
					_, err = fileHandle.Write(body)
					if err != nil {
						log.Printf("Err write file, %s", err)
						continue
					}

					fileHandle.Sync()
					fileHandle.Close()
				}
			}
		}
	}
}

// создаем док из картинок
func createDoc(config ConfigFile, web ConfigWeb) string {

	var filePath string
	counterImg := 1 // counter name photo

	for _, project := range config.Projects {

		// search name project
		if project.Name != web.Project {
			continue
		}

		//создаем папку, где будем хранить наши отчётики
		createFolder(fmt.Sprintf("reports/%s", project.Name))

		// Load default values from configs field
		configs := project.Configs

		// прогружаем теплейт, если не получиться, то просто создаём новый
		doc, err := document.OpenTemplate("template/template.docx")
		if err != nil {
			log.Printf("error opening Windows Word 2016 document: %s", err)
			doc = document.New()
		}

		// Handle every instruction
		for _, instr := range project.Instructions {
			// Check is there instructions for text, add paragraph if true
			text := instr.TextJson.Content
			if text != "" {
				style := instr.TextJson.Style
				para := doc.AddParagraph()
				para.SetStyle(style)
				run := para.AddRun()
				run.AddText(text)
				//run.AddField(document.)
			}

			// Check is there instructions for grafana, then add images and description to doc
			dash := instr.Grafana
			if dash.PanelsID != nil {
				reg := regexp.MustCompile(`(.*/|^)(.+?/.+?)(\?|$)`)
				dashboard := getOrDefault(dash.Dashboard, configs.Grafana.Dashboard)
				dPath := reg.FindStringSubmatch(dashboard.(string))

				for _, panelId := range dash.PanelsID {
					// формируем адрес картинки
					filePath = fmt.Sprintf("imgs/%s/%s/%s.png", project.Name, strings.Split(dPath[2], "/")[1], panelId)

					// открытие картинки
					img, err := common.ImageFromFile(filePath)
					if err != nil {
						log.Printf("Error doc1, unable to create image(%s): %s", err, filePath)
						continue
					}

					// добавляем картинку, получаем ссылку на картинку
					iref, err := doc.AddImage(img)
					if err != nil {
						log.Printf("Error doc2, unable to add image (%s) to document: %s", err, filePath)
						continue
					}

					// добавляем параграф с названием картинки
					para := doc.AddParagraph()
					para.SetStyle(getOrDefault(dash.DescStyle, configs.Grafana.DescStyle).(string))
					run := para.AddRun()
					run.AddText(fmt.Sprintf("Рисунок "))
					run.AddField(" SEQ Рисунок \\* ARABIC ")
					run.AddText(fmt.Sprintf(". %s", getOrDefault(dash.Description, configs.Grafana.Description)))

					// добавляем параграф с фото
					para = doc.AddParagraph()

					imgInl, err := para.AddRun().AddDrawingInline(iref)
					if err != nil {
						log.Printf("Error doc3, unable to add inline image(%s): %s", err, filePath)
						continue
					}

					//Width, _ := strconv.ParseFloat(getOrDefault(dash.Width, configs.Grafana.Width).(string), 64)
					//Height, _ := strconv.ParseFloat(getOrDefault(dash.Height, configs.Grafana.Height).(string), 64)
					//
					//Width /= 63.8
					//Height /= 57.1

					Width := 16
					Height := 9

					imgInl.SetSize(measurement.Centimeter*measurement.Distance(Width), measurement.Centimeter*measurement.Distance(Height))

					counterImg++
				}
			}
		}
		reportLink = fmt.Sprintf("reports/%s/express-%d-%02d-%02d_%02d-%02d-%02d.docx", project.Name, time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(), time.Now().Minute(), time.Now().Second())
		doc.SaveToFile(reportLink)
	}
	return reportLink
}

func createFolder(name string) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		os.MkdirAll(name, 0666)
	}
}

func convertTime(times string, timeZone int64) string {
	var timeStr string
	strArr := regexp.MustCompile(`(\d+)[./-](\d+)[./-](\d+) (\d+):(\d+:\d+)`).FindStringSubmatch(times) // select item from row
	// 1 month
	// 2 day
	// 3 years
	// 4 hours
	// 5 minute and second

	i, err := strconv.ParseInt(strArr[4], 10, 64)
	if err != nil {
		log.Printf("Error time, %s", err)
		return ""
	}

	if i >= 12 {
		// PM
		fmt.Sprintf("%s %d:%s PM", strArr[1], i-12, strArr[3])
		timeStr = fmt.Sprintf("%s/%s/%s %d:%s PM", strArr[2], strArr[1], strArr[3], i-12, strArr[5])
	} else {
		// AM
		timeStr = fmt.Sprintf("%s/%s/%s %d:%s AM", strArr[2], strArr[1], strArr[3], i, strArr[5])
	}

	layout := "01/02/2006 3:04:05 PM"
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		log.Printf("Error time, %s", err)
		return ""
	}
	// умножаем на 1000 т.к в графане время в милисекундах
	// вычитаем 10800 т.е. 3 часа для свдига в часовую зону МСК
	return fmt.Sprintf("%d", (t.Unix()-timeZone)*1000)
}

//28/01/2018 16:30:05
func upload(w http.ResponseWriter, r *http.Request) {

	// delete folder with file TODO debug
	//err := os.RemoveAll("imgs")
	//if err != nil {
	//	log.Printf("Error delete folder %s", err)
	//}

	// Config program
	configurationFile := ConfigFile{}
	configurationFile.init("tsconfig.json")

	if r.Method == "GET" {
		// текущее время
		configurationFile.Times = fmt.Sprintf(time.Now().Format("02-01-2006 15:04:05"))
		t, _ := template.ParseFiles("template/upload.gtpl")
		t.Execute(w, configurationFile)
	} else {
		configurationWeb := ConfigWeb{r.FormValue("timeTo"), r.FormValue("timeFrom"), r.FormValue("project"), r.FormValue("timezone")}
		// real work
		getInfo(configurationFile, configurationWeb)
		// collate img
		reportLink = createDoc(configurationFile, configurationWeb) // получаем ссылку на отчёт

		// add ABS path
		dir, err := filepath.Abs(filepath.Dir(reportLink))
		if err != nil {
			log.Println(err)
		}
		reportLink = dir + "\\" + filepath.Base(reportLink)

		// work with html
		t, _ := template.ParseFiles("template/result.gtpl")
		t.Execute(w, struct {
			ReportLink string
		}{reportLink})
	}
}

func main() {
	// configurator for logger
	var str = "log" // name folder for logs

	// check what folder log is exist
	_, err := os.Stat(str)
	if os.IsNotExist(err) {
		os.MkdirAll(str, 0666)
	}
	str = fmt.Sprintf("%s/%d-%02d-%02d-%02d-%02d-%02d-logFile.log", str, time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(), time.Now().Minute(), time.Now().Second())
	// open a file
	f, err := os.OpenFile(str, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()

	// assign it to the standard logger
	log.SetOutput(f) // TODO config logs
	log.SetPrefix("generateDoc ")

	// start server
	http.HandleFunc("/", upload) // setting router rule

	port := flag.String("port", "9005", "the port value")
	flag.Parse()

	log.Printf("Start work. Port: %s", *port)
	defer log.Println("Stop work.")

	err = http.ListenAndServe(":"+*port, nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	log.Println("Start work!")
}
