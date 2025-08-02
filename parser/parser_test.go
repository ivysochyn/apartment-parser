package parser

import (
	"regexp"
	"testing"
)

func TestExtractOffer(t *testing.T) {
	// Test with the actual HTML example
	sampleHTML := `<div data-cy="l-card" data-testid="l-card" class="css-1sw7q4x">
		<div class="css-1apmciz">
			<div data-cy="ad-card-title" class="css-u2ayx9">
				<a class="css-1tqlkj0" href="/d/oferta/wynajme-kawalerke-na-osiedlu-przy-ul-cukrowej-w-szczecinie-CID3-ID16SdDt.html">
					<h4 class="css-1g61gc2">Wynajmę kawalerkę na osiedlu przy ul. Cukrowej w Szczecinie</h4>
				</a>
				<p data-testid="ad-price" class="css-uj7mm0">1 700 zł<span class="css-18rym86"></span></p>
			</div>
			<div class="css-odp1qd">
				<p data-testid="location-date" class="css-vbz67q">Szczecin, Gumieńce - Dzisiaj o 14:30</p>
			</div>
		</div>
	</div>`

	offer := extractOffer(sampleHTML)

	// Verify extracted data
	if offer.Title != "Wynajmę kawalerkę na osiedlu przy ul. Cukrowej w Szczecinie" {
		t.Errorf("Title not extracted correctly: got %q", offer.Title)
	}

	if offer.Price != 1700 {
		t.Errorf("Price not extracted correctly: got %d, want 1700", offer.Price)
	}

	if offer.Location != "Szczecin, Gumieńce" {
		t.Errorf("Location not extracted correctly: got %q", offer.Location)
	}

	if offer.Url != "https://www.olx.pl/d/oferta/wynajme-kawalerke-na-osiedlu-przy-ul-cukrowej-w-szczecinie-CID3-ID16SdDt.html" {
		t.Errorf("URL not extracted correctly: got %q", offer.Url)
	}

	// Time should be adjusted for timezone (14:30 + 2 hours = 16:30)
	if offer.Time != "16:30" {
		t.Errorf("Time not extracted/adjusted correctly: got %q, want 16:30", offer.Time)
	}
}

func TestExtractOfferWithCustomConfig(t *testing.T) {
	// Test with a custom configuration to show flexibility
	customConfig := ExtractorConfig{
		TitleSelector: Selector{
			Tag:       "h3",  // Different tag
			Attribute: "",
			Value:     "",
		},
		PriceSelector: Selector{
			Tag:       "span",
			Attribute: "class",
			Value:     "price",
		},
		LocationSelector: Selector{
			Tag:       "div",
			Attribute: "class",
			Value:     "location",
		},
		URLSelector: Selector{
			Tag:       "a",
			Attribute: "href",
			Value:     "",
		},
		PricePattern:   regexp.MustCompile(`\d+`),
		TimePattern:    regexp.MustCompile(`\d{2}:\d{2}`),
		TodayKeyword:   "Today",
		BaseURL:        "https://example.com",
		TimezoneOffset: 0,
	}

	customHTML := `<div>
		<a href="/offer/123">
			<h3>Test Apartment</h3>
		</a>
		<span class="price">2500 PLN</span>
		<div class="location">Warsaw - Today at 10:00</div>
	</div>`

	offer := extractOfferWithConfig(customHTML, customConfig)

	if offer.Title != "Test Apartment" {
		t.Errorf("Custom config: Title not extracted correctly: got %q", offer.Title)
	}

	if offer.Price != 2500 {
		t.Errorf("Custom config: Price not extracted correctly: got %d", offer.Price)
	}
}

func TestExtractPrice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Simple price", "1500 zł", 1500},
		{"Price with space", "1 700 zł", 1700},
		{"Price with non-breaking space", "2\u00a0000 PLN", 2000},
		{"Multiple numbers", "Price: 3000 zł/month", 3000},
		{"No price", "Contact for price", 0},
	}

	pattern := regexp.MustCompile(`\d+`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPrice(tt.input, pattern)
			if result != tt.expected {
				t.Errorf("extractPrice(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractLocationAndTime(t *testing.T) {
	config := OLXConfig

	tests := []struct {
		name         string
		input        string
		wantLocation string
		wantTime     string
	}{
		{
			name:         "Today with time",
			input:        "Warszawa, Mokotów - Dzisiaj o 14:30",
			wantLocation: "Warszawa, Mokotów",
			wantTime:     "16:30", // +2 hours timezone
		},
		{
			name:         "Not today",
			input:        "Kraków - 02 sierpnia 2025",
			wantLocation: "Kraków",
			wantTime:     "", // Should be empty as it's not today
		},
		{
			name:         "Location only",
			input:        "Gdańsk",
			wantLocation: "Gdańsk",
			wantTime:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, timeStr := extractLocationAndTime(tt.input, config)
			if location != tt.wantLocation {
				t.Errorf("Location = %q, want %q", location, tt.wantLocation)
			}
			if timeStr != tt.wantTime {
				t.Errorf("Time = %q, want %q", timeStr, tt.wantTime)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		baseURL string
		want    string
	}{
		{
			"Relative URL",
			"/d/oferta/test.html",
			"https://www.olx.pl",
			"https://www.olx.pl/d/oferta/test.html",
		},
		{
			"Absolute URL",
			"https://www.olx.pl/d/oferta/test.html",
			"https://www.olx.pl",
			"https://www.olx.pl/d/oferta/test.html",
		},
		{
			"Path without slash",
			"oferta/test.html",
			"https://www.olx.pl",
			"https://www.olx.pl/oferta/test.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeURL(tt.url, tt.baseURL)
			if got != tt.want {
				t.Errorf("normalizeURL(%q, %q) = %q, want %q", tt.url, tt.baseURL, got, tt.want)
			}
		})
	}
}
