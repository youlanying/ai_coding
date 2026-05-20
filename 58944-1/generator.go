package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type Generator struct {
	outputDir string
	prefix    string
}

func NewGenerator(outputDir, prefix string) *Generator {
	return &Generator{
		outputDir: outputDir,
		prefix:    prefix,
	}
}

func (g *Generator) GenerateAll(data *ExcelData) error {
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := g.generateConfig(data); err != nil {
		return fmt.Errorf("generate config error: %w", err)
	}

	if err := g.generatePayTable(data); err != nil {
		return fmt.Errorf("generate pay_table error: %w", err)
	}

	if err := g.generatePrize(data); err != nil {
		return fmt.Errorf("generate prize error: %w", err)
	}

	if err := g.generateBaseReelStrip(data); err != nil {
		return fmt.Errorf("generate base_reel_strip error: %w", err)
	}

	if err := g.generateFreeReelStrip(data); err != nil {
		return fmt.Errorf("generate free_reel_strip error: %w", err)
	}

	return nil
}

func (g *Generator) generateConfig(data *ExcelData) error {
	if len(data.Config) == 0 {
		return nil
	}

	keys := make([]string, 0, len(data.Config))
	for k := range data.Config {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sorted := make(map[string]ConfigItem)
	for _, k := range keys {
		sorted[k] = data.Config[k]
	}

	return g.writeJSON("config", sorted)
}

func (g *Generator) generatePayTable(data *ExcelData) error {
	if len(data.PayTable) == 0 {
		return nil
	}

	result := make(map[string]map[string]interface{})

	keys := make([]string, 0, len(data.PayTable))
	for k := range data.PayTable {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ni, _ := strconv.Atoi(keys[i])
		nj, _ := strconv.Atoi(keys[j])
		return ni < nj
	})

	for _, k := range keys {
		item := data.PayTable[k]
		entry := make(map[string]interface{})
		entry["id"] = item.ID

		payKeys := make([]string, 0, len(item.Pay))
		for pk := range item.Pay {
			payKeys = append(payKeys, pk)
		}
		sort.Slice(payKeys, func(i, j int) bool {
			ni, _ := strconv.Atoi(payKeys[i])
			nj, _ := strconv.Atoi(payKeys[j])
			return ni < nj
		})

		for _, pk := range payKeys {
			entry[pk] = item.Pay[pk]
		}

		result[k] = entry
	}

	return g.writeJSON("pay_table", result)
}

func (g *Generator) generatePrize(data *ExcelData) error {
	if len(data.Prize) == 0 {
		return nil
	}

	keys := make([]string, 0, len(data.Prize))
	for k := range data.Prize {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ni, _ := strconv.Atoi(keys[i])
		nj, _ := strconv.Atoi(keys[j])
		return ni < nj
	})

	sorted := make(map[string]PrizeItem)
	for _, k := range keys {
		sorted[k] = data.Prize[k]
	}

	return g.writeJSON("prize", sorted)
}

func (g *Generator) generateBaseReelStrip(data *ExcelData) error {
	if len(data.BaseReelStrip) == 0 {
		return nil
	}

	keys := make([]string, 0, len(data.BaseReelStrip))
	for k := range data.BaseReelStrip {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ni, _ := strconv.Atoi(keys[i])
		nj, _ := strconv.Atoi(keys[j])
		return ni < nj
	})

	sorted := make(map[string]ReelStripItem)
	for _, k := range keys {
		sorted[k] = data.BaseReelStrip[k]
	}

	return g.writeJSON("base_reel_strip", sorted)
}

func (g *Generator) generateFreeReelStrip(data *ExcelData) error {
	if len(data.FreeReelStrip) == 0 {
		return nil
	}

	keys := make([]string, 0, len(data.FreeReelStrip))
	for k := range data.FreeReelStrip {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ni, _ := strconv.Atoi(keys[i])
		nj, _ := strconv.Atoi(keys[j])
		return ni < nj
	})

	sorted := make(map[string]ReelStripItem)
	for _, k := range keys {
		sorted[k] = data.FreeReelStrip[k]
	}

	return g.writeJSON("free_reel_strip", sorted)
}

func (g *Generator) writeJSON(suffix string, data interface{}) error {
	fileName := fmt.Sprintf("%s_%s.json", g.prefix, suffix)
	filePath := filepath.Join(g.outputDir, fileName)

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", " ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("encode json error: %w", err)
	}

	formatted := g.formatJSON(buf.String())

	if err := os.WriteFile(filePath, []byte(formatted), 0644); err != nil {
		return fmt.Errorf("write file error: %w", err)
	}

	fmt.Printf("Generated: %s\n", filePath)
	return nil
}

func (g *Generator) formatJSON(input string) string {
	var data interface{}
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return input
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", " ")
	_ = encoder.Encode(data)

	result := buf.String()
	result = result[:len(result)-1]
	return result
}
