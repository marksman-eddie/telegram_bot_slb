package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	goquery "github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type File struct {
	id   string
	Name string
	Link string
}

type Row struct {
	Name        string
	Revision    string
	Type        string
	Description string
	Files       []File
}

func DownloadFile(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return bytes
}

func DownloadFileToExplore(fileName string, url string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fileErr := ioutil.WriteFile(fileName, bytes, 0644)
	if fileErr != nil {
		panic(err)
	}

}

func ress(url string, search string) []Row {
	var rowList []Row
	req, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()
	doc, err := goquery.NewDocumentFromReader(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("tbody").Each(func(index int, tbody *goquery.Selection) {
		if tbody.Children().Length() > 100 {
			tbody.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
				var newRow = new(Row)
				searchTrue := false
				rowhtml.Find("td").EachWithBreak(func(indextd int, tablecell *goquery.Selection) bool {
					if strings.Contains(tablecell.Text(), search) {
						searchTrue = true
						return false
					}
					return true
				})
				if searchTrue {
					rowhtml.Find("tbody").Each(func(indexbody int, resRow *goquery.Selection) {
						resRow.Find("a").Each(func(indexa int, link *goquery.Selection) {
							href, ok := link.Attr("href")
							if ok {
								File := File{strconv.Itoa(indexa), link.Text(), url + href}
								newRow.Files = append(newRow.Files, File)
							}
						})
					})
					rowhtml.Find("td").Each(func(writeIndexTd int, writeTablecell *goquery.Selection) {
						if writeTablecell.Children().Length() < 1 {
							if writeIndexTd == 0 {
								newRow.Name = writeTablecell.Text()
							}
							if writeIndexTd == 1 {
								newRow.Revision = writeTablecell.Text()
							}
							if writeIndexTd == 2 {
								newRow.Type = writeTablecell.Text()
							}
							if writeIndexTd == 3 {
								newRow.Description = writeTablecell.Text()
							}
						}
					})
					rowList = append(rowList, *newRow)
				}
			})
		}
	})
	return rowList
}
func Contains(a []File, x string) bool {
	for _, n := range a {
		if x == n.id {
			return true
		}
	}
	return false
}

func main() {
	bot, err := tgbotapi.NewBotAPI("your_api_key")
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if !update.Message.IsCommand() {
			continue
		}
		switch update.Message.Command() {
		case "find_document":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			msg.Text = "Please enter the name or description for the document you want to search..."
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
			for update := range updates {
				if update.Message != nil {
					url := "https://docs.gems.slb.com/ENG/Mid-Tier_Pump/"
					linkList := ress(url, update.Message.Text)
					if len(linkList) == 0 {
						msg.Text = "I don't know that name or description. \nTry it again... \n\t /find_document - search document in GEMS"
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
					} else {
						var printFileList []File
						for i := 0; i < len(linkList); i++ {
							for j := 0; j < len(linkList[i].Files); j++ {
								printFile := File{
									id:   strconv.Itoa(i + j + 1),
									Name: linkList[i].Files[j].Name,
									Link: linkList[i].Files[j].Link,
								}
								printFileList = append(printFileList, printFile)
								msg.Text = printFile.id + "\n\nFile name:\n" + linkList[i].Files[j].Name + "\n\nFile description:\n" + linkList[i].Description + "\n\nRevision:\n" + linkList[i].Revision + "\n\nType:\n" + linkList[i].Type
								if _, err := bot.Send(msg); err != nil {
									log.Panic(err)
								}
							}
						}
						for updateM := range updates {
							if update.Message == nil {
								continue
							}
							if updateM.Message != nil {
								if updateM.Message.Text == "0" {
									break
									msg.Text = "You could use :\n\t /find_document - search document in GEMS"
									if _, err := bot.Send(msg); err != nil {
										log.Panic(err)
									}
								} else {
									if Contains(printFileList, updateM.Message.Text) {
										for _, v := range printFileList {
											if v.id == updateM.Message.Text {
												resp, err := http.Get(v.Link)
												if err != nil {
													panic(err)
												}
												bytes, err := ioutil.ReadAll(resp.Body)
												if err != nil {
													panic(err)
												}
												FileBytes := tgbotapi.FileBytes{
													Name:  v.Name,
													Bytes: bytes,
												}

												if _, err := bot.Send(tgbotapi.NewDocument(int64(updateM.Message.Chat.ID), FileBytes)); err != nil {
													log.Panic(err)
												}

											}
										}
										msg.Text = "Enter 0 for break"
										if _, err := bot.Send(msg); err != nil {
											log.Panic(err)
										}

									} else {
										msg.Text = "Wrong file number!"
										if _, err := bot.Send(msg); err != nil {
											log.Panic(err)
										}
									}

								}

							}

						}

						//}
					}
					break
				} else {
					msg.Text = "I don't know that name or description. \nTry it again... \n\t /open - search document in GEMS"
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					break
				}
			}

		case "start":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! This is telegram bot will help you find the document on NAME or Description. \nYou could use :\n\t /find_document - search document in GEMS")
			bot.Send(msg)

		case "gif":
			//https://upload.wikimedia.org/wikipedia/ru/archive/6/6b/20210505175821%21NyanCat.gif
			resp, err := http.Get("https://upload.wikimedia.org/wikipedia/ru/archive/6/6b/20210505175821%21NyanCat.gif")
			if err != nil {
				panic(err)
			}
			bytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}
			FileBytes := tgbotapi.FileBytes{
				Name:  "picture.gif",
				Bytes: bytes,
			}
			if _, err := bot.Send(tgbotapi.NewAnimation(int64(update.Message.Chat.ID), FileBytes)); err != nil {
				log.Panic(err)
			}
		case "open":
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			msg.Text = "Please enter the name or description for the document you want to search..."
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
			for update := range updates {
				if update.Message != nil {
					url := "https://docs.gems.slb.com/ENG/Mid-Tier_Pump/"
					linkList := ress(url, update.Message.Text)
					if len(linkList) == 0 {
						msg.Text = "I don't know that name or description. \nTry it again... \n\t /find_document - search document in GEMS"
						if _, err := bot.Send(msg); err != nil {
							log.Panic(err)
						}
					} else {
						for i := 0; i < len(linkList); i++ {
							for j := 0; j < len(linkList[i].Files); j++ {
								resp, err := http.Get(linkList[i].Files[j].Link)
								if err != nil {
									panic(err)
								}
								bytes, err := ioutil.ReadAll(resp.Body)
								if err != nil {
									panic(err)
								}
								FileBytes := tgbotapi.FileBytes{
									Name:  linkList[i].Name,
									Bytes: bytes,
								}

								if _, err := bot.Send(tgbotapi.NewDocument(int64(update.Message.Chat.ID), FileBytes)); err != nil {
									log.Panic(err)
								}

								//DownloadFile(linkList[i].Name)
								//fmt.Println(linkList[i].Name + " created")
								//FileBytes := tgbotapi.FileBytes{
								//	Name:  linkList[i].Name,
								//	Bytes: DownloadFile(url + linkList[i].Name),
								//}
								//if _, err := bot.Send(tgbotapi.NewDocument(int64(update.Message.Chat.ID), FileBytes)); err != nil {
								//	log.Panic(err)
								//}

							}
						}
					}
					break
				} else {
					msg.Text = "I don't know that name or description. \nTry it again... \n\t /find_document - search document in GEMS"
					if _, err := bot.Send(msg); err != nil {
						log.Panic(err)
					}
					break
				}
			}

		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			msg.Text = "I don't know that command"
			if _, err := bot.Send(msg); err != nil {
				log.Panic(err)
			}
		}
		//if _, err := bot.Send(msg); err != nil {
		//	log.Panic(err)
		//}
	}
}
