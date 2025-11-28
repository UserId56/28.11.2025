package services

import (
	"io"
	"log"

	"25.11.2025/models"
	"github.com/signintech/gopdf"
)

func PDFGeneration(w io.Writer, data models.TaskResponse) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	err := pdf.AddTTFFont("Arial", "ArialMT.ttf")
	if err != nil {
		log.Print(err.Error())
		return err
	}
	err = pdf.SetFont("Arial", "", 14)
	if err != nil {
		log.Print(err.Error())
		return err
	}

	pdf.AddHeader(func() {
		pdf.SetY(20)
		pdf.SetX(175)
		pdf.Cell(nil, "Отчет о доступности ресурсов")
	})

	pdf.AddPage()
	tableStartY := 75.0
	tableStartX := 150.0
	table := pdf.NewTableLayout(tableStartX, tableStartY, 30, len(data.Links))
	table.AddColumn("Ресурс", 150, "center")
	table.AddColumn("Статус", 150, "center")
	for link, status := range data.Links {
		table.AddRow([]string{link, status})
	}
	err = table.DrawTable()
	if err != nil {
		log.Print(err.Error())
		return err
	}
	_, err = pdf.WriteTo(w)
	if err != nil {
		log.Print(err.Error())
		return err
	}
	return nil
}
