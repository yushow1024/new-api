package model

import "testing"

func TestGenerationContentType(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: "/v1/videos/generations", want: "video"},
		{path: "/v1/videos/generations/", want: "video"},
		{path: "/v1/images/edits", want: "image"},
		{path: "/v1/images/edits/", want: "image"},
		{path: "/v1/images/generations", want: "image"},
		{path: "/v1/images/generations/", want: "image"},
		{path: "/v1/chat/completions", want: ""},
		{path: "/v1/videos", want: ""},
		{path: "/v1/videos/generations-extra", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := generationContentType(tt.path); got != tt.want {
				t.Fatalf("generationContentType(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
