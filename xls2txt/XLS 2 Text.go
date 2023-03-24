/*
File Name:  XLS 2 Text.go
Copyright:  2019 Kleissner Investments s.r.o.
Author:     Peter Kleissner

The code originally used https://github.com/extrame/xls, which revealed multiple bugs that crashed for certain Excel files.
Now it forks the xls package and the underlying ole2 package. This fork also fixes excessive memory usage issues.
*/

package xls2txt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/FleetWiser/fileconversion/xls"
)

// XLS2Text extracts text from an Excel sheet. It returns bytes written.
// The parameter size is the max amount of bytes (not characters) to write out.
// The whole Excel file is required even for partial text extraction. This function returns no error with 0 bytes written in case of corrupted or invalid file.
func XLS2Text(reader io.ReadSeeker, writer io.Writer, size int64) (written int64, err error) {

	xlFile, err := xls.OpenReader(reader, "utf-8")
	if err != nil || xlFile == nil {
		return 0, err
	}

	for n := 0; n < xlFile.NumSheets(); n++ {
		if sheet1 := xlFile.GetSheet(n); sheet1 != nil {
			if err = writeOutput(writer, []byte(xlGenerateSheetTitle(sheet1.Name, n, int(sheet1.MaxRow))), &written, &size); err != nil || size == 0 {
				return written, err
			}

			for m := 0; m <= int(sheet1.MaxRow); m++ {
				row1 := sheet1.Row(m)
				if row1 == nil {
					continue
				}

				rowText := ""

				// go through all columns
				for c := row1.FirstCol(); c < row1.LastCol(); c++ {
					if text := row1.Col(c); text != "" {
						text = cleanCell(text)

						if c > row1.FirstCol() {
							rowText += ", "
						}
						rowText += text
					}
				}

				rowText += "\n"

				if err = writeOutput(writer, []byte(rowText), &written, &size); err != nil || size == 0 {
					return written, err
				}
			}
		}
	}

	return written, nil
}

// XLS2CSV converts selected sheet of the XLS file into CSV format.
func XLS2CSV(reader io.ReadSeeker, sheetNumber int) ([]byte, error) {
	xlFile, err := xls.OpenReader(reader, "utf-8")
	if err != nil || xlFile == nil {
		return nil, err
	}

	sheet := xlFile.GetSheet(sheetNumber)
	if nil == sheet {
		return nil, errors.New("sheet doesn't exist")
	}

	rows := make([]string, 0)
	for ii := 0; ii < int(sheet.MaxRow); ii++ {
		row := sheet.Row(ii)
		if row == nil {
			continue
		}

		columns := make([]string, 0)
		for jj := row.FirstCol(); jj < row.LastCol(); jj++ {
			columns = append(columns, WrapCSVCell(row.Col(jj)))
		}

		rows = append(rows, strings.Join(columns, ","))
	}

	return []byte(strings.Join(rows, "\n")), nil
}

// cleanCell returns a cleaned cell text without new-lines
func cleanCell(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.TrimSpace(text)

	return text
}

func WrapCSVCell(cell string) string {
	return "\"" + cleanCell(cell) + "\""
}

func xlGenerateSheetTitle(name string, number, rows int) (title string) {
	if number > 0 {
		title += "\n"
	}

	title += fmt.Sprintf("Sheet \"%s\" (%d rows):\n", name, rows)

	return title
}

func writeOutput(writer io.Writer, output []byte, alreadyWritten *int64, size *int64) (err error) {

	if int64(len(output)) > *size {
		output = output[:*size]
	}

	*size -= int64(len(output))

	writtenOut, err := writer.Write(output)
	*alreadyWritten += int64(writtenOut)

	return err
}

// IsFileXLS checks if the data indicates a XLS file
// XLS has a signature of D0 CF 11 E0 A1 B1 1A E1
func IsFileXLS(data []byte) bool {
	return bytes.HasPrefix(data, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})
}

// XLS2Cells converts an XLS file to individual cells
func XLS2Cells(reader io.ReadSeeker) (cells []string, err error) {

	xlFile, err := xls.OpenReader(reader, "utf-8")
	if err != nil || xlFile == nil {
		return nil, err
	}

	for n := 0; n < xlFile.NumSheets(); n++ {
		if sheet1 := xlFile.GetSheet(n); sheet1 != nil {
			for m := 0; m <= int(sheet1.MaxRow); m++ {
				row1 := sheet1.Row(m)
				if row1 == nil {
					continue
				}

				for c := row1.FirstCol(); c < row1.LastCol(); c++ {
					if text := row1.Col(c); text != "" {
						text = cleanCell(text)
						cells = append(cells, text)
					}
				}
			}
		}
	}

	return
}
