package repository

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mrscorpio/uahelper/internal/tagdata"
)

func StoreData(d *tagdata.AllTags, arhDirName string, periodic bool) (*bytes.Buffer, string, error) {
	data, err := json.Marshal(*d)
	if err != nil {
		return nil, "", err
	}
	nowT := time.Now()
	prevHourT := nowT.Add(-1 * time.Hour)

	filename := ""

	if periodic {
		filename = prevHourT.Format("20060102_15")
	} else {
		filename = "stop_" + nowT.Format("0102_1504")
	}

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create(filename + ".json")
	if err != nil {
		return nil, "", err
	}
	_, err = f.Write(data)
	if err != nil {
		return nil, "", err
	}
	w.Close()

	err = os.WriteFile(arhDirName+filename, buf.Bytes(), 0755)

	if err != nil {
		return nil, "", err
	} else if periodic {
		err := os.Remove(arhDirName + prevHourT.Format("20060102_15") + ".json")
		if err != nil {
			return nil, "", err
		}
	}
	/*
		if cfg.Bot && !periodic {
			prms := &bot.SendDocumentParams{
				ChatID:   cfg.BotChat,
				Document: &models.InputFileUpload{Filename: filename, Data: bytes.NewReader(buf.Bytes())},
				Caption:  "прога для просмотра: https://disk.yandex.ru/d/P3LXkuUmBDBTtA",
			}
			b.SendDocument(ctx, prms)
		}
	*/
	return buf, filename, nil
}

func ReadStored(d *tagdata.AllTags, filename string) (time.Time, error) {
	tm := time.Now()
	r, err := zip.OpenReader(filename)
	if err != nil {
		return tm, err
	}
	defer r.Close()

	dateFromFile := strings.Split(filename, "_")
	if len(dateFromFile) > 2 {
		tm, err = time.Parse("0102_1504", dateFromFile[1]+"_"+dateFromFile[2])
		if err != nil {
			log.Println(err)
		}
	}

	buf := new(bytes.Buffer)
	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return tm, err
		}
		n, err := buf.ReadFrom(rc)
		if err != nil {
			return tm, err
		}
		fmt.Println(n)
		rc.Close()
	}
	err = json.Unmarshal(buf.Bytes(), d)
	if err != nil {
		return tm, err
	}
	return tm, nil
}
