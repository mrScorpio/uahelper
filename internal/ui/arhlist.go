package ui

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mrscorpio/uahelper/configs"
	"github.com/mrscorpio/uahelper/internal/tagdata"
)

func SearchDataFiles() ([]string, error) {
	fileList := make([]string, 0)
	arhFiles, err := os.ReadDir("./arh")
	if err == nil {
		for _, f := range arhFiles {
			if strings.HasPrefix(f.Name(), "stop_") {
				fileList = append(fileList, "arh/"+f.Name())
			}
		}
	} else {
		log.Println(err)
	}

	homeFiles, err := os.ReadDir(".")
	if err != nil {
		return fileList, nil
	}

	for _, f := range homeFiles {
		if strings.HasPrefix(f.Name(), "stop_") {
			fileList = append(fileList, f.Name())
		}
	}

	return fileList, nil
}

func HttPath(d *tagdata.AllTags, cfg *configs.Config, ipaddr string) string {
	commaTags := ""
	for k, v := range cfg.ShowTags {
		if commaTags != "" {
			commaTags += ","
		}
		if v {
			commaTags += d.Tag[k].Name
		}
	}
	return "http://" + ipaddr + cfg.TrPort +
		"/?show=" + commaTags + "&zoom=" + d.TripTag +
		"&begin=" + fmt.Sprint(FstInd) + "&end=" + fmt.Sprint(LastInd) +
		"&step=" + fmt.Sprint((LastInd-FstInd)/1000)
}
