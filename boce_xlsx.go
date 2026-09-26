package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func writeRowsXLSX(path string, rows []DetectRow) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	files := map[string]string{
		"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
</Types>`,
		"_rels/.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`,
		"xl/workbook.xml": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="检测结果" sheetId="1" r:id="rId1"/>
  </sheets>
</workbook>`,
		"xl/_rels/workbook.xml.rels": `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
</Relationships>`,
		"xl/worksheets/sheet1.xml": buildWorksheetXML(rows),
	}

	order := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"xl/workbook.xml",
		"xl/_rels/workbook.xml.rels",
		"xl/worksheets/sheet1.xml",
	}
	for _, name := range order {
		if err := addZipString(zipWriter, name, files[name]); err != nil {
			return err
		}
	}

	return nil
}

func defaultExportFilename(rows []DetectRow) string {
	detectType := "检测结果"
	if len(rows) > 0 && strings.TrimSpace(rows[0].Type) != "" {
		detectType = strings.TrimSpace(rows[0].Type)
	}
	return fmt.Sprintf("%s_%s.xlsx", sanitizeFilenamePart(detectType), time.Now().Format("20060102"))
}

func sanitizeFilenamePart(value string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		`"`, "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	value = strings.TrimSpace(replacer.Replace(value))
	if value == "" {
		return "检测结果"
	}
	return value
}

func addZipString(zipWriter *zip.Writer, name string, content string) error {
	writer, err := zipWriter.Create(name)
	if err != nil {
		return err
	}
	_, err = io.WriteString(writer, content)
	return err
}

func buildWorksheetXML(rows []DetectRow) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	if isICPRows(rows) {
		writeXLSXRow(&builder, 1, []string{"域名", "是否备案", "备案号", "网站名称"})
		for idx, row := range rows {
			writeXLSXRow(&builder, idx+2, []string{
				firstNonEmpty(row.Domain, row.Target),
				row.Status,
				row.BeianCode,
				row.SiteName,
			})
		}
		builder.WriteString(`</sheetData></worksheet>`)
		return builder.String()
	}

	writeXLSXRow(&builder, 1, []string{"类型", "检测目标", "状态", "检测时间", "错误 / 备注"})
	for idx, row := range rows {
		writeXLSXRow(&builder, idx+2, []string{
			row.Type,
			row.Target,
			row.Status,
			row.CheckedAt,
			row.ErrorRemark,
		})
	}
	builder.WriteString(`</sheetData></worksheet>`)
	return builder.String()
}

func isICPRows(rows []DetectRow) bool {
	return len(rows) > 0 && rows[0].Type == TypeICP
}

func writeXLSXRow(builder *strings.Builder, rowNum int, values []string) {
	builder.WriteString(fmt.Sprintf(`<row r="%d">`, rowNum))
	for idx, value := range values {
		cellRef := fmt.Sprintf("%s%d", xlsxColumnName(idx), rowNum)
		builder.WriteString(fmt.Sprintf(`<c r="%s" t="inlineStr"><is><t>`, cellRef))
		escapeXML(builder, value)
		builder.WriteString(`</t></is></c>`)
	}
	builder.WriteString(`</row>`)
}

func xlsxColumnName(idx int) string {
	name := ""
	for idx >= 0 {
		name = string(rune('A'+idx%26)) + name
		idx = idx/26 - 1
	}
	return name
}

func escapeXML(builder *strings.Builder, value string) {
	value = stripInvalidXMLChars(value)
	if value == "" {
		return
	}
	_ = xml.EscapeText(builder, []byte(value))
}

// stripInvalidXMLChars 剔除 XML 1.0 Char 之外的字符，保留 \t \n \r。
// XML 1.0 的合法字符是 #x9 | #xA | #xD | [#x20-#xD7FF] | [#xE000-#xFFFD] | [#x10000-#x10FFFF]，
// 其余（0x00-0x08、0x0B、0x0C、0x0E-0x1F、代理区、U+FFFE/U+FFFF 等）进入单元格
// 会产出 Excel 打不开的非法 xlsx，这里在转义前先删掉。
func stripInvalidXMLChars(value string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == 0x09 || r == 0x0A || r == 0x0D:
			return r
		case r >= 0x20 && r <= 0xD7FF:
			return r
		case r >= 0xE000 && r <= 0xFFFD:
			return r
		case r >= 0x10000 && r <= 0x10FFFF:
			return r
		default:
			return -1
		}
	}, value)
}
