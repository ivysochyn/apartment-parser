package telegrambot

import (
	"apartment-parser/parser"
	"strings"
	"testing"
)

func TestProcessPriceStr(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMin     int
		wantMax     int
		wantErr     bool
		errContains string
	}{
		// Range format tests
		{
			name:    "valid range",
			input:   "1000-2000",
			wantMin: 1000,
			wantMax: 2000,
			wantErr: false,
		},
		{
			name:    "valid range with spaces",
			input:   " 1000 - 2000 ",
			wantMin: 1000,
			wantMax: 2000,
			wantErr: false,
		},
		{
			name:        "invalid range - min higher than max",
			input:       "2000-1000",
			wantErr:     true,
			errContains: "minimum price cannot be higher than maximum price",
		},

		// Minimum only tests
		{
			name:    "minimum only with plus",
			input:   "1500+",
			wantMin: 1500,
			wantMax: 0,
			wantErr: false,
		},
		{
			name:    "minimum only with dash",
			input:   "1500-",
			wantMin: 1500,
			wantMax: 0,
			wantErr: false,
		},
		{
			name:    "minimum only with plus and spaces",
			input:   " 1500 + ",
			wantMin: 1500,
			wantMax: 0,
			wantErr: false,
		},
		{
			name:        "invalid minimum with plus",
			input:       "abc+",
			wantErr:     true,
			errContains: "invalid minimum price format",
		},
		{
			name:        "negative minimum with plus",
			input:       "-500+",
			wantErr:     true,
			errContains: "minimum price cannot be negative",
		},

		// Maximum only tests
		{
			name:    "maximum only with dash prefix",
			input:   "-2000",
			wantMin: 0,
			wantMax: 2000,
			wantErr: false,
		},
		{
			name:    "maximum only - just number",
			input:   "2000",
			wantMin: 0,
			wantMax: 2000,
			wantErr: false,
		},
		{
			name:    "maximum only with spaces",
			input:   " 2000 ",
			wantMin: 0,
			wantMax: 2000,
			wantErr: false,
		},
		{
			name:        "invalid maximum with dash",
			input:       "-abc",
			wantErr:     true,
			errContains: "invalid maximum price format",
		},
		{
			name:        "zero maximum",
			input:       "-0",
			wantErr:     true,
			errContains: "maximum price must be positive",
		},
		{
			name:    "negative maximum as number",
			input:   "-500",
			wantErr: false,
			wantMin: 0,
			wantMax: 500,
		},

		// Edge cases
		{
			name:    "zero minimum with range",
			input:   "0-1000",
			wantMin: 0,
			wantMax: 1000,
			wantErr: false,
		},
		{
			name:        "empty input",
			input:       "",
			wantErr:     true,
			errContains: "invalid price format",
		},
		{
			name:        "only dash",
			input:       "-",
			wantErr:     true,
			errContains: "invalid maximum price format",
		},
		{
			name:        "multiple dashes",
			input:       "1000-2000-3000",
			wantErr:     true,
			errContains: "invalid price range format",
		},
		{
			name:        "invalid characters",
			input:       "abc-def",
			wantErr:     true,
			errContains: "invalid minimum price",
		},
		{
			name:        "zero as single number",
			input:       "0",
			wantErr:     true,
			errContains: "price must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotMax, err := processPriceStr(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("processPriceStr() expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("processPriceStr() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("processPriceStr() unexpected error = %v", err)
				return
			}

			if gotMin != tt.wantMin {
				t.Errorf("processPriceStr() gotMin = %v, want %v", gotMin, tt.wantMin)
			}
			if gotMax != tt.wantMax {
				t.Errorf("processPriceStr() gotMax = %v, want %v", gotMax, tt.wantMax)
			}
		})
	}
}

func TestPriceRangeToURLIntegration(t *testing.T) {
	tests := []struct {
		name           string
		priceInput     string
		city           string
		wantURLContains []string
		wantURLNotContains []string
	}{
		{
			name:       "full range",
			priceInput: "1000-2000",
			city:       "poznan",
			wantURLContains: []string{
				"poznan",
				"search[filter_float_price:from]=1000",
				"search[filter_float_price:to]=2000",
			},
		},
		{
			name:       "minimum only with plus",
			priceInput: "1500+",
			city:       "warszawa",
			wantURLContains: []string{
				"warszawa",
				"search[filter_float_price:from]=1500",
			},
			wantURLNotContains: []string{
				"search[filter_float_price:to]",
			},
		},
		{
			name:       "maximum only with dash",
			priceInput: "-3000",
			city:       "krakow",
			wantURLContains: []string{
				"krakow",
				"search[filter_float_price:to]=3000",
			},
			wantURLNotContains: []string{
				"search[filter_float_price:from]",
			},
		},
		{
			name:       "maximum only as number",
			priceInput: "2500",
			city:       "gdansk",
			wantURLContains: []string{
				"gdansk",
				"search[filter_float_price:to]=2500",
			},
			wantURLNotContains: []string{
				"search[filter_float_price:from]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse price
			minPrice, maxPrice, err := processPriceStr(tt.priceInput)
			if err != nil {
				t.Fatalf("processPriceStr() error = %v", err)
			}

			// Create search term
			searchTerm := parser.SearchTerm{
				Location:  tt.city,
				Price_min: float64(minPrice),
				Price_max: float64(maxPrice),
			}

			// Create URL
			url, err := parser.CreateUrl(searchTerm)
			if err != nil {
				t.Fatalf("CreateUrl() error = %v", err)
			}

			// Check URL contains expected strings
			for _, want := range tt.wantURLContains {
				if !strings.Contains(url, want) {
					t.Errorf("URL = %v, want to contain %v", url, want)
				}
			}

			// Check URL doesn't contain unwanted strings
			for _, notWant := range tt.wantURLNotContains {
				if strings.Contains(url, notWant) {
					t.Errorf("URL = %v, should not contain %v", url, notWant)
				}
			}
		})
	}
}
