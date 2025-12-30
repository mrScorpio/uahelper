package repository

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mrscorpio/uahelper/internal/tagdata"
)

func StoreData(d *tagdata.AllTags, arhDirName string, periodic bool) error {
	data, err := json.Marshal(*d)
	if err != nil {
		return err
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
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	w.Close()

	err = os.WriteFile(arhDirName+filename+".zip", buf.Bytes(), 0755)

	if err != nil {
		return err
	} else if periodic {
		err := os.Remove(arhDirName + prevHourT.Format("20060102_15") + ".json")
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadStored(d *tagdata.AllTags, filename string) error {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer r.Close()
	buf := new(bytes.Buffer)
	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return err
		}
		n, err := buf.ReadFrom(rc)
		if err != nil {
			return err
		}
		fmt.Println(n)
		rc.Close()
	}
	err = json.Unmarshal(buf.Bytes(), d)
	if err != nil {
		return err
	}
	return nil
}
