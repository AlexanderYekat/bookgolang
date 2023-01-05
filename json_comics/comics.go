package main

import (
	xkcd "comics/basejsons"
	"comics/s3general"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	//s3yandex.InitCheckS3yandex()
	CreateBaseOfAllComics(5)
	spComics, err := s3general.GetComics("asstudiotest", "test", "test")
	if err != nil {
		fmt.Printf("ошибка поиска комикса %v\n", err)
	} else {
		fmt.Printf("резульатт поиска %v\n", spComics)
	}
}
func findComics(text string) (string, string, error) {
	arrxkcd, arrbyteBuf, err := s3general.FindComics(text)
	fmt.Println(arrxkcd, arrbyteBuf, err)
	return "URL", "transcriptions", nil
}
func CreateBaseOfAllComics(max int) error {
	lastNumb, err := GetLastNumber()
	if err != nil {
		return fmt.Errorf("ошибка получения последнего номера комикса %v\n", err)
	}
	for i := max; i <= lastNumb; i++ {
		res, err := http.Get(fmt.Sprintf("https://xkcd.com/%d/info.0.json", i))
		defer res.Body.Close()
		if err != nil {
			if BreakeCycle(i, fmt.Sprintf("получение ответа с сайта xkcd (%v)", err)) {
				return err
			}
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			sErr := fmt.Sprintf("чтения тела ответа (%v)", err)
			if BreakeCycle(i, sErr) {
				return err
			}
		}
		var comics xkcd.TComics
		err = json.Unmarshal(body, &comics)
		if err != nil {
			if BreakeCycle(i, fmt.Sprintf("разбор json ответа (%v)", err)) {
				return err
			}
		}
		fmt.Println("---begin yandex ----")
		err = s3general.WriteComics(comics, "asstudiotest")
		if err != nil {
			fmt.Printf("ошибка добавления записи комикса в хранилище s3 %v\n", err)
		}
		fmt.Println("---end yandex ------")
		WriteComics(i, comics)
		if i >= max {
			fmt.Println("достигнут максимум по комиксам")
			break
		}
	} //цикл по всем комиксам сайта
	return nil
}
func WriteComics(num int, comics xkcd.TComics) error {
	err := xkcd.WriteInFile(comics)
	if err != nil {
		fmt.Errorf("Ошибка %v записи в файл комикса с номером %d\n", err, num)
		return err
	}
	return nil
}
func GetLastNumber() (int, error) {
	res, err := http.Get("https://xkcd.com/info.0.json")
	if err != nil {
		return 0, fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("ошибка чтения тела ответа: %v", err)
	}
	var comics xkcd.TComics
	err = json.Unmarshal(body, &comics)
	if err != nil {
		return 0, fmt.Errorf("ошибка преобразования json %v", err)
	}
	return comics.Num, nil
}
func BreakeCycle(num int, strErr string) bool {
	fmt.Errorf("Ошибка %s на комиксе c номером %d\n", strErr, num)
	fmt.Println("Проолжить? (y/n(defaut:y)")
	ans := ""
	_, err := fmt.Scan(ans)
	if err == nil {
		if ans == "n" {
			return true
		}
	}
	return false
}
