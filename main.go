package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	awsarn "github.com/aws/aws-sdk-go-v2/aws/arn"
)

func main() {
	// CLI flags
	inputPath := flag.String("input", "", "Path to input AWS Configuration Item JSON file")
	outputPath := flag.String("output", "", "Optional path to save anonymized JSON (default: stdout)")
	dryRun := flag.Bool("dry-run", false, "Print anonymized JSON to stdout without saving")
	pretty := flag.Bool("pretty", true, "Pretty-print JSON output")

	flag.Parse()

	if *inputPath == "" {
		fmt.Println("Error: --input is required")
		flag.Usage()
		os.Exit(1)
	}

	// Read file
	data, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Printf("Failed to read input file: %v\n", err)
		os.Exit(1)
	}

	// Decode JSON (support single object or array)
	var input any
	if err := json.Unmarshal(data, &input); err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		os.Exit(1)
	}

	// Anonymize
	anonymized := anonymize(input)

	// Re-encode
	var out []byte
	if *pretty {
		out, err = json.MarshalIndent(anonymized, "", "  ")
	} else {
		out, err = json.Marshal(anonymized)
	}
	if err != nil {
		fmt.Printf("Failed to encode anonymized JSON: %v\n", err)
		os.Exit(1)
	}

	if *dryRun || *outputPath == "" {
		fmt.Println(string(out))
	} else {
		err := os.WriteFile(*outputPath, out, 0644)
		if err != nil {
			fmt.Printf("Failed to write output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Anonymized JSON written to: %s\n", *outputPath)
	}
}

func anonymize(data any) any {
	switch v := data.(type) {
	case []any:
		for i, item := range v {
			v[i] = anonymize(item)
		}
		return v
	case map[string]any:
		return anonymizeMap(v)
	case string:
		if awsarn.IsARN(v) {
			return anonymizeArn(v)
		}
		return v
	default:
		return data
	}
}

func anonymizeMap(ci map[string]any) map[string]any {
	redacted := make(map[string]any)
	for k, v := range ci {
		lowerK := strings.ToLower(k)

		vs, ok := v.(string)
		if ok && awsarn.IsARN(vs) {
			v = anonymizeArn(vs)
		}

		if strings.Contains(lowerK, "name") {
			redacted[k] = "REDACTED_NAME"
			continue
		}

		switch lowerK {
		case "accountid":
			redacted[k] = "000000000000"
		case "resourceid":
			redacted[k] = "REDACTED_RESOURCE_ID"
		case "resourcename":
			redacted[k] = "REDACTED_RESOURCE_NAME"
		case "arn":
			vs, ok := v.(string)
			if ok && strings.Contains(vs, "REDACTED") {
				redacted[k] = vs
			} else if ok {
				redacted[k] = anonymizeArn(vs)
			} else {
				redacted[k] = "arn:aws:REDACTED"
			}
			redacted[k] = "arn:aws:REDACTED"
		case "tags":
			redacted[k] = map[string]string{"REDACTED": "REDACTED"}
		case "configuration":
			if submap, ok := v.(map[string]any); ok {
				redacted[k] = anonymize(submap)
			} else {
				redacted[k] = "REDACTED_CONFIGURATION"
			}
		case "relationships":
			if arr, ok := v.([]any); ok {
				for i, item := range arr {
					if m, ok := item.(map[string]any); ok {
						arr[i] = anonymizeMap(m)
					}
				}
				redacted[k] = arr
			} else {
				redacted[k] = v
			}
		default:
			// Recurse if value is map or list
			switch val := v.(type) {
			case map[string]any:
				redacted[k] = anonymize(val)
			case []any:
				redacted[k] = anonymize(val)
			default:
				redacted[k] = v
			}
		}
	}
	return redacted
}

func anonymizeArn(arn string) string {
	arnParts, err := awsarn.Parse(arn)
	if err != nil {
		return "arn:aws:REDACTED"
	}
	arnParts.AccountID = "000000000000"
	arnParts.Resource = "REDACTED"
	return arnParts.String()
}
