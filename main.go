package main

import (
	"fmt"
	"html/template"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	// Folder path for repo.
	folder := "./data/go"
	days := 60

	// Logrus.
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	// Calculate Repository activity index and vitality.
	vitalityIndex, vitality, err := CalculateRepoActivity(folder, days)
	if err != nil {
		log.Errorf("error calculating repository Activity to file: %v", err)
	}
	log.Debugf("Activity Index from Today() for %s is %f", folder, vitalityIndex)

	// Prepare the js slice.
	type JSData struct {
		VitalitySlice []int
	}
	jsData := JSData{}
	for i := 0; i < len(vitality); i++ {
		jsData.VitalitySlice = append(jsData.VitalitySlice, int(vitality[i]))
	}

	t, err := template.ParseFiles("template.tpl")
	if err != nil {
		panic(err)
	}
	f, err := os.OpenFile("index.html", os.O_WRONLY, 0777)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	err = t.ExecuteTemplate(f, "template.tpl", jsData)
	if err != nil {
		panic(err)
	}
}
