package main

import "testing"

func TestLanguageForTag(t *testing.T) {
	tests := []struct {
		tag  string
		want uiLanguage
	}{
		{tag: "de", want: german},
		{tag: "de-DE", want: german},
		{tag: "en-US", want: english},
		{tag: "fr-FR", want: english},
		{tag: "", want: english},
	}

	for _, test := range tests {
		if got := languageForTag(test.tag); got != test.want {
			t.Errorf("languageForTag(%q) = %q, want %q", test.tag, got, test.want)
		}
	}
}

func TestTextsForLanguage(t *testing.T) {
	if got := textsFor(german).statusReady; got != "Bereit." {
		t.Errorf("German ready text = %q, want %q", got, "Bereit.")
	}
	if got := textsFor(english).statusReady; got != "Ready." {
		t.Errorf("English ready text = %q, want %q", got, "Ready.")
	}
}
