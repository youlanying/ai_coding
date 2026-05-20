package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type ConfigItem struct {
	Name        string        `json:"name"`
	StringValue string        `json:"string_value"`
	IntValue    int           `json:"int_value"`
	ArrayValue  []interface{} `json:"array_value"`
}

type PayTableItem struct {
	ID int            `json:"id"`
	Pay map[string]int `json:"-"`
}

type PrizeItem struct {
	ID        int    `json:"id"`
	Prize     int    `json:"prize"`
	Type      int    `json:"type"`
	JackpotID int    `json:"jackpot_id"`
}

type ReelStripItem struct {
	ID     int           `json:"id"`
	ReelID int           `json:"reel_id"`
	Col    int           `json:"col"`
	Strip  [][]int       `json:"strip"`
}

type ExcelData struct {
	Config        map[string]ConfigItem
	PayTable      map[string]PayTableItem
	Prize         map[string]PrizeItem
	BaseReelStrip map[string]ReelStripItem
	FreeReelStrip map[string]ReelStripItem
}

type Parser struct {
	filePath string
	f        *excelize.File
}

func NewParser(filePath string) *Parser {
	return &Parser{filePath: filePath}
}

func (p *Parser) Parse() (*ExcelData, error) {
	var err error
	p.f, err = excelize.OpenFile(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer p.f.Close()

	data := &ExcelData{
		Config:        make(map[string]ConfigItem),
		PayTable:      make(map[string]PayTableItem),
		Prize:         make(map[string]PrizeItem),
		BaseReelStrip: make(map[string]ReelStripItem),
		FreeReelStrip: make(map[string]ReelStripItem),
	}

	sheets := p.f.GetSheetList()
	for _, sheet := range sheets {
		normalized := strings.ToLower(sheet)
		normalized = strings.TrimPrefix(normalized, "#<")
		normalized = strings.TrimSuffix(normalized, ">")

		switch normalized {
		case "config":
			if err := p.parseConfig(sheet, data); err != nil {
				return nil, fmt.Errorf("parse config sheet error: %w", err)
			}
		case "pay_table", "paytable":
			if err := p.parsePayTable(sheet, data); err != nil {
				return nil, fmt.Errorf("parse pay_table sheet error: %w", err)
			}
		case "prize":
			if err := p.parsePrize(sheet, data); err != nil {
				return nil, fmt.Errorf("parse prize sheet error: %w", err)
			}
		case "base_reel_strip", "basereelstrip", "base_reel":
			if err := p.parseReelStrip(sheet, data, true); err != nil {
				return nil, fmt.Errorf("parse base_reel_strip sheet error: %w", err)
			}
		case "free_reel_strip", "freereelstrip", "free_reel":
			if err := p.parseReelStrip(sheet, data, false); err != nil {
				return nil, fmt.Errorf("parse free_reel_strip sheet error: %w", err)
			}
		}
	}

	return data, nil
}

func (p *Parser) parseConfig(sheet string, data *ExcelData) error {
	rows, err := p.f.GetRows(sheet)
	if err != nil {
		return err
	}

	if len(rows) < 2 {
		return nil
	}

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 || row[0] == "" {
			continue
		}

		name := strings.TrimSpace(row[0])
		item := ConfigItem{
			Name:        name,
			StringValue: "",
			IntValue:    0,
			ArrayValue:  []interface{}{},
		}

		if len(row) > 1 {
			item.StringValue = strings.TrimSpace(row[1])
		}

		if len(row) > 2 && row[2] != "" {
			if intVal, err := strconv.Atoi(strings.TrimSpace(row[2])); err == nil {
				item.IntValue = intVal
			}
		}

		if len(row) > 3 && row[3] != "" {
			arrayStr := strings.TrimSpace(row[3])
			item.ArrayValue = p.parseConfigArray(arrayStr, name)
		}

		data.Config[name] = item
	}

	return nil
}

func (p *Parser) parseConfigArray(s string, name string) []interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return []interface{}{}
	}

	s = p.unquoteStrings(s)

	if strings.Contains(s, "],[") {
		if !strings.HasPrefix(s, "[") {
			s = "[" + s
		}
		if !strings.HasSuffix(s, "]") {
			s = s + "]"
		}

		inner := s[1 : len(s)-1]
		if strings.Contains(inner, "],[") && strings.Contains(inner, "[") && strings.Contains(inner, "]") {
			firstPair := strings.SplitN(inner, "],[", 2)[0]
			if strings.Count(firstPair, "[") > 1 || strings.Contains(firstPair, "],[") {
				return p.parseArrayString(s)
			}
		}

		parts := p.splitNestedPairs(s)
		result := []interface{}{}
		for _, pair := range parts {
			pair = strings.TrimSpace(pair)
			pair = strings.TrimPrefix(pair, "[")
			pair = strings.TrimSuffix(pair, "]")

			if strings.Contains(pair, "[") && strings.Contains(pair, "]") {
				nestedResult := p.parseArrayString("[" + pair + "]")
				if len(nestedResult) > 0 {
					result = append(result, nestedResult)
				}
				continue
			}

			nums := strings.Split(pair, ",")
			pairResult := []interface{}{}
			for _, num := range nums {
				num = strings.TrimSpace(num)
				if val, err := strconv.Atoi(num); err == nil {
					pairResult = append(pairResult, val)
				} else if val, err := strconv.ParseFloat(num, 64); err == nil {
					pairResult = append(pairResult, val)
				}
			}
			if len(pairResult) > 0 {
				result = append(result, pairResult)
			}
		}
		return result
	}

	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		inner := s[1 : len(s)-1]
		if !strings.Contains(inner, "[") && !strings.Contains(inner, "]") {
			parts := strings.Split(inner, ",")
			result := []interface{}{}
			for _, part := range parts {
				part = strings.TrimSpace(part)
				part = strings.Trim(part, "\"")
				if val, err := strconv.Atoi(part); err == nil {
					result = append(result, val)
				} else if val, err := strconv.ParseFloat(part, 64); err == nil {
					result = append(result, val)
				} else if part != "" {
					result = append(result, part)
				}
			}
			return []interface{}{result}
		}
		return p.parseArrayString(s)
	}

	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		result := []interface{}{}
		for _, part := range parts {
			part = strings.TrimSpace(part)
			part = strings.Trim(part, "\"")
			if val, err := strconv.Atoi(part); err == nil {
				result = append(result, val)
			} else if val, err := strconv.ParseFloat(part, 64); err == nil {
				result = append(result, val)
			} else if part != "" {
				result = append(result, part)
			}
		}
		return result
	}

	if val, err := strconv.Atoi(s); err == nil {
		return []interface{}{val}
	}

	s = strings.Trim(s, "\"")
	return []interface{}{s}
}

func (p *Parser) splitNestedPairs(s string) []string {
	result := []string{}
	depth := 0
	current := ""
	for _, ch := range s {
		switch ch {
		case '[':
			if depth == 0 && current != "" {
				current = ""
			}
			depth++
			current += string(ch)
		case ']':
			depth--
			current += string(ch)
			if depth == 0 {
				result = append(result, strings.TrimSpace(current))
				current = ""
			}
		case ',':
			if depth == 0 {
				continue
			}
			current += string(ch)
		default:
			current += string(ch)
		}
	}
	if depth > 0 && strings.TrimSpace(current) != "" {
		result = append(result, strings.TrimSpace(current))
	}
	return result
}

func (p *Parser) unquoteStrings(s string) string {
	result := ""
	inQuote := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' {
			inQuote = !inQuote
			continue
		}
		result += string(ch)
	}
	return result
}

func (p *Parser) parseCommaArray(s string) []interface{} {
	result := []interface{}{}
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if val, err := strconv.Atoi(part); err == nil {
			result = append(result, val)
		} else if val, err := strconv.ParseFloat(part, 64); err == nil {
			result = append(result, val)
		} else {
			result = append(result, part)
		}
	}
	return result
}

func (p *Parser) parsePayTable(sheet string, data *ExcelData) error {
	rows, err := p.f.GetRows(sheet)
	if err != nil {
		return err
	}

	if len(rows) < 3 {
		return nil
	}

	headers := rows[0]
	cleanHeaders := make([]string, len(headers))
	for j, h := range headers {
		h = strings.TrimSpace(h)
		h = strings.TrimSuffix(h, "K")
		cleanHeaders[j] = h
	}

	for i := 2; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 || row[0] == "" {
			continue
		}

		id, _ := strconv.Atoi(row[0])
		item := PayTableItem{
			ID:  id,
			Pay: make(map[string]int),
		}

		for j := 1; j < len(cleanHeaders) && j < len(row); j++ {
			if cleanHeaders[j] == "" || row[j] == "" {
				continue
			}
			if val, err := strconv.Atoi(row[j]); err == nil {
				item.Pay[cleanHeaders[j]] = val
			}
		}

		data.PayTable[row[0]] = item
	}

	return nil
}

func (p *Parser) parsePrize(sheet string, data *ExcelData) error {
	rows, err := p.f.GetRows(sheet)
	if err != nil {
		return err
	}

	if len(rows) < 2 {
		return nil
	}

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 || row[0] == "" {
			continue
		}

		id, _ := strconv.Atoi(row[0])
		item := PrizeItem{
			ID:        id,
			Prize:     0,
			Type:      0,
			JackpotID: 0,
		}

		if len(row) > 1 && row[1] != "" {
			item.Prize, _ = strconv.Atoi(strings.TrimSpace(row[1]))
		}
		if len(row) > 3 && row[3] != "" {
			item.Type, _ = strconv.Atoi(strings.TrimSpace(row[3]))
		}
		if len(row) > 4 && row[4] != "" {
			item.JackpotID, _ = strconv.Atoi(strings.TrimSpace(row[4]))
		}

		data.Prize[row[0]] = item
	}

	return nil
}

func (p *Parser) parseReelStrip(sheet string, data *ExcelData, isBase bool) error {
	rows, err := p.f.GetRows(sheet)
	if err != nil {
		return err
	}

	if len(rows) < 3 {
		return nil
	}

	for i := 2; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 || row[0] == "" {
			continue
		}

		id, _ := strconv.Atoi(row[0])
		item := ReelStripItem{
			ID:     id,
			ReelID: 0,
			Col:    0,
			Strip:  [][]int{},
		}

		if len(row) > 1 && row[1] != "" {
			item.ReelID, _ = strconv.Atoi(row[1])
		}
		if len(row) > 2 && row[2] != "" {
			item.Col, _ = strconv.Atoi(row[2])
		}
		if len(row) > 3 && row[3] != "" {
			item.Strip = p.parseStripArray(row[3])
		}

		if isBase {
			data.BaseReelStrip[row[0]] = item
		} else {
			data.FreeReelStrip[row[0]] = item
		}
	}

	return nil
}

func (p *Parser) parseArrayString(s string) []interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return []interface{}{}
	}

	result := []interface{}{}
	
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = s[1 : len(s)-1]
		result = p.parseNestedArray(s)
	} else {
		parts := strings.Split(s, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if val, err := strconv.Atoi(part); err == nil {
				result = append(result, val)
			} else if val, err := strconv.ParseFloat(part, 64); err == nil {
				result = append(result, val)
			} else {
				result = append(result, part)
			}
		}
	}

	return result
}

func (p *Parser) parseNestedArray(s string) []interface{} {
	result := []interface{}{}
	depth := 0
	current := ""
	items := []string{}

	for _, ch := range s {
		switch ch {
		case '[':
			depth++
			current += string(ch)
		case ']':
			depth--
			current += string(ch)
		case ',':
			if depth == 0 {
				items = append(items, strings.TrimSpace(current))
				current = ""
			} else {
				current += string(ch)
			}
		default:
			current += string(ch)
		}
	}
	if current != "" {
		items = append(items, strings.TrimSpace(current))
	}

	for _, item := range items {
		item = strings.TrimSpace(item)
		if strings.HasPrefix(item, "[") && strings.HasSuffix(item, "]") {
			inner := item[1 : len(item)-1]
			nested := p.parseNestedArray(inner)
			result = append(result, nested)
		} else {
			if val, err := strconv.Atoi(item); err == nil {
				result = append(result, val)
			} else if val, err := strconv.ParseFloat(item, 64); err == nil {
				result = append(result, val)
			} else if item != "" {
				result = append(result, strings.Trim(item, "\"'"))
			}
		}
	}

	return result
}

func (p *Parser) parseStripArray(s string) [][]int {
	result := [][]int{}
	s = strings.TrimSpace(s)
	if s == "" {
		return result
	}

	if !strings.HasPrefix(s, "[") {
		s = "[" + s
	}
	if !strings.HasSuffix(s, "]") {
		s = s + "]"
	}

	arr := p.parseArrayString(s)
	
	for _, item := range arr {
		switch v := item.(type) {
		case []interface{}:
			pair := []int{}
			for _, num := range v {
				switch n := num.(type) {
				case int:
					pair = append(pair, n)
				case float64:
					pair = append(pair, int(n))
				}
			}
			if len(pair) > 0 {
				result = append(result, pair)
			}
		case int:
			result = append(result, []int{v})
		case float64:
			result = append(result, []int{int(v)})
		}
	}

	return result
}
