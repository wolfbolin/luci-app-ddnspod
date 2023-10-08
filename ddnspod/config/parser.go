package config

import (
	"os"
	"strings"
)

func ParserConfig(configPath string) (*UciConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	config := string(data)
	config = strings.ReplaceAll(config, "\r\n", "\n")
	configLines := strings.Split(config, "\n")

	configMap := UciConfig{Section: make(map[string]UciSection)}
	var tmp *UciSection
	for _, line := range configLines {
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) <= 0 {
			continue
		}

		fieldVal := strings.Trim(fields[2], "'")
		if fields[0] == "option" {
			tmp.Option[fields[1]] = fieldVal
		} else if fields[0] == "list" {
			tmp.List[fields[1]] = append(tmp.List[fields[1]], fieldVal)
		} else if fields[0] == "config" {
			if tmp != nil {
				configMap.Section[tmp.Key] = *tmp
			}

			sectionKey := "default"
			if len(fields) >= 3 {
				sectionKey = fieldVal
			}

			tmp = &UciSection{
				Key:    sectionKey,
				Type:   fields[1],
				List:   make(map[string]UciList),
				Option: make(map[string]UciOption),
			}
		}
	}
	if tmp != nil {
		configMap.Section[tmp.Key] = *tmp
	}

	return &configMap, nil
}
