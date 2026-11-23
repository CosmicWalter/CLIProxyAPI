package executor

import (
	"bytes"
	"mime/multipart"
	"testing"

	"github.com/MoeForge/moego/ojson"
)

func TestCodexImageGIFParameters(t *testing.T) {
	payload := []byte(`{"model":"gpt-image-2.6-flare","prompt":"A looping animation","output_format":"gif","frame_count":4,"frame_duration_ms":150}`)
	for _, action := range []string{"generate", "edit", "multipart"} {
		t.Run(action, func(t *testing.T) {
			var prepared codexOpenAIImagePreparedRequest
			var err error
			switch action {
			case "generate":
				prepared, err = codexPrepareOpenAIImageGenerationJSON(payload, "gpt-image-2.6-flare")
			case "edit":
				prepared, err = codexPrepareOpenAIImageEditJSON(payload, "gpt-image-2.6-flare")
			case "multipart":
				var body bytes.Buffer
				writer := multipart.NewWriter(&body)
				for key, value := range map[string]string{"model": "gpt-image-2.6-flare", "prompt": "A looping animation", "output_format": "gif", "frame_count": "4", "frame_duration_ms": "150"} {
					if errWrite := writer.WriteField(key, value); errWrite != nil {
						t.Fatal(errWrite)
					}
				}
				if errClose := writer.Close(); errClose != nil {
					t.Fatal(errClose)
				}
				prepared, err = codexPrepareOpenAIImageEditMultipart(body.Bytes(), "gpt-image-2.6-flare", writer.FormDataContentType())
			}
			if err != nil {
				t.Fatal(err)
			}
			for field, want := range map[string]int64{"frame_count": 4, "frame_duration_ms": 150} {
				value := ojson.GetBytes(prepared.Body, "tools.0."+field)
				if value.Type != ojson.Number || value.Int() != want {
					t.Fatalf("%s = %s, want numeric %d", field, value.Raw, want)
				}
			}
			if got := ojson.GetBytes(prepared.Body, "tools.0.output_format").String(); got != "gif" {
				t.Fatalf("output format = %q", got)
			}
		})
	}
}

func TestCodexImageGIFMultipartValuesAreNumeric(t *testing.T) {
	for _, field := range []string{"frame_count", "frame_duration_ms"} {
		got := codexOpenAIImageEditFormJSONValue(field, "150")
		if string(got) != "150" {
			t.Fatalf("%s = %s, want number", field, got)
		}
	}
}

func TestCodexImageGIFMimeType(t *testing.T) {
	if got := codexMimeTypeFromOutputFormat(" GIF "); got != "image/gif" {
		t.Fatalf("MIME = %q", got)
	}
}

func TestCodexImageGIFResponsePreservesAnimation(t *testing.T) {
	const encodedGIF = "R0lGODlhAQABAAAAACwAAAAAAQABAAA="
	image := codexImageCallResult{Result: encodedGIF, OutputFormat: "gif"}
	for _, format := range []string{"url", "b64_json"} {
		body, err := codexBuildImagesAPIResponse([]codexImageCallResult{image}, 123, nil, image, format)
		if err != nil {
			t.Fatal(err)
		}
		if got := ojson.GetBytes(body, "output_format").String(); got != "gif" {
			t.Fatalf("format = %q", got)
		}
		field, want := "data.0.b64_json", encodedGIF
		if format == "url" {
			field, want = "data.0.url", "data:image/gif;base64,"+encodedGIF
		}
		if got := ojson.GetBytes(body, field).String(); got != want {
			t.Fatalf("GIF response changed: %q", got)
		}
	}
}

func TestCodexImageGIFDirectMultipartParameters(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for field, value := range map[string]string{"output_format": "gif", "frame_count": "4", "frame_duration_ms": "150"} {
		if err := writer.WriteField(field, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	payload, _, err := codexPrepareDirectOpenAIImageEditPayload(body.Bytes(), "gpt-image-2.6-flare", writer.FormDataContentType(), false)
	if err != nil {
		t.Fatal(err)
	}
	for field, want := range map[string]int64{"frame_count": 4, "frame_duration_ms": 150} {
		got := ojson.GetBytes(payload, field)
		if got.Type != ojson.Number || got.Int() != want {
			t.Fatalf("%s = %s, want numeric %d", field, got.Raw, want)
		}
	}
	if got := ojson.GetBytes(payload, "output_format").String(); got != "gif" {
		t.Fatalf("format = %q", got)
	}
}
