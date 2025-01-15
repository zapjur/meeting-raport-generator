package main

import (
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"os"
	"path/filepath"
)

func generatePDF(meetingID string, transcriptions []Transcription, summary Summary, ocrResults []OCRResult, screenshots []string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "B", 16)

	pdf.AddPage()
	pdf.Cell(40, 10, fmt.Sprintf("Meeting Report - %s", meetingID))

	pdf.SetFont("Arial", "", 12)
	pdf.Ln(10)
	pdf.Cell(0, 10, "Summary:")
	pdf.Ln(10)
	pdf.MultiCell(0, 10, summary.SummaryText, "", "", false)

	pdf.Ln(10)
	pdf.Cell(0, 10, "Transcriptions:")
	pdf.Ln(10)
	for _, t := range transcriptions {
		pdf.Cell(0, 10, fmt.Sprintf("[%s-%s] %s: %s", t.TimestampStart, t.TimestampEnd, t.SpeakerID, t.Transcription))
		pdf.Ln(5)
	}

	pdf.Ln(10)
	pdf.Cell(0, 10, "OCR Results:")
	pdf.Ln(10)
	for _, ocr := range ocrResults {
		pdf.MultiCell(0, 10, ocr.TextResult, "", "", false)
		pdf.Ln(5)
	}

	pdf.Ln(10)
	pdf.Cell(0, 10, "Screenshots:")
	pdf.Ln(10)
	for _, screenshot := range screenshots {
		pdf.Image(screenshot, 10, pdf.GetY(), 180, 0, false, "", 0, "")
		pdf.Ln(95)
	}

	outputDir := filepath.Join("/shared-report", meetingID)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	outputPath := filepath.Join(outputDir, fmt.Sprintf("meeting_report_%s.pdf", meetingID))

	err := pdf.OutputFileAndClose(outputPath)
	if err != nil {
		return fmt.Errorf("error generating PDF: %v", err)
	}

	return nil
}
