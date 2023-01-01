// функиця по хранению json данных и их изсвлечению
// для различных баз данных
package xkcd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

const dirfilesbase = "filesbase"

type TComics = struct {
	Transcript string `json:"transcript"`
	Num        int    `json:"num"`
	Month      string `json:"month"`
	Link       string `json:"link"`
	News       string `json:"news"`
	SafeTitle  string `json:"safe_title"`
	Alt        string `json:"alt"`
	Img        string `json:"img"`
	Title      string `json:"title"`
	Day        string `json:"day"`
}

// просто запись в файл
func WriteInFile(comics TComics) error {
	if _, err := os.Stat(dirfilesbase); os.IsNotExist(err) {
		os.Mkdir(dirfilesbase, 0744)
	}
	strNum := strconv.Itoa(comics.Num)
	fileName := strNum + ".json"
	f, err := os.OpenFile("./"+dirfilesbase+"/"+fileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Ошбика открытия файла %v (%v). Комикс с номером %d - не записан в базу c ошибкой.\n", fileName, err, comics.Num)
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")
	err = enc.Encode(comics)
	if err != nil {
		fmt.Printf("Ошибка (%v) записи комикcа %d в файл.\n", err, comics.Num)
		return err
	}
	err = writeJPGComics(comics.Num, comics.Img)
	if err != nil {
		fmt.Printf("Ошибка %v получение картинки комикса %v", err, comics.Num)
	}
	return nil
}

// запись картинки
func writeJPGComics(num int, addrrUrl string) error {
	resp, err := http.Get(addrrUrl)
	if err != nil {
		fmt.Printf("Ошибка %v загурзки каринки коммикса %v по адресу %v\n", err, num, addrrUrl)
		return err
	}
	strNum := strconv.Itoa(num)
	fileName := strNum + ".jpg"
	fullFileName := "./" + dirfilesbase + "/" + fileName
	f, err := os.OpenFile(fullFileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Ошибка %v открытия файла %v для записи картинки комикса %v\n", err, fullFileName, num)
		return err
	}
	defer f.Close()
	io.Copy(f, resp.Body)
	resp.Body.Close()
	return nil
}
func grepFile() {
}
func writeCouchDB() {
}
func getFromCouchDB() {
}
func writeMongoDB() {
}
func getFromMongoDB() {
}
func writeTextSearchManticore() {
}
func getTextSearchManticore() {
}
func writeInOpenSearch() {
}
func getFromOpenSeearch() {
}
func writeS3() {

}
func getS3() {
}
