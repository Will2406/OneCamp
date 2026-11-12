package models

import (
	"encoding/json"
	"strings"
	"testing"
)

// A doc's metadata write carries its space to its own files' entries (a Docs &
// Files file is its own doc), so a file moved to another project is found by
// that project's people and no longer by the old one's. A file in a comment
// keeps no space, as it never had one, and a doc taken out of every space
// takes the space off its files.
func TestADocsMetadataWriteCarriesItsSpaceToItsFiles(t *testing.T) {
	title, private := "logo.png", true
	parse := func(space string) (source string, params map[string]any, query string) {
		var body struct {
			Script struct {
				Source string         `json:"source"`
				Params map[string]any `json:"params"`
			} `json:"script"`
			Query json.RawMessage `json:"query"`
		}
		raw := docMetadataBody("f1", &title, &private, []string{"r"}, []string{"e"}, []string{"c"}, "u", space)
		if err := json.Unmarshal([]byte(raw), &body); err != nil {
			t.Fatalf("body isn't JSON: %v\n%s", err, raw)
		}
		return body.Script.Source, body.Script.Params, string(body.Query)
	}
	source, params, query := parse("p2")
	if params["space"] != "p2" {
		t.Errorf("space param %v; want p2", params["space"])
	}
	const ownFiles = "if (ctx._index == 'attachments' && ctx._source.attachment_comment_id == null) { if (params.space == '') { ctx._source.remove('attachment_doc_space_uuid'); } else { ctx._source.attachment_doc_space_uuid = params.space; } }"
	if !strings.Contains(source, ownFiles) {
		t.Errorf("script %q; want it to set the space on the doc's own files only", source)
	}
	if strings.Count(source, "attachment_doc_space_uuid") != 2 {
		t.Errorf("script %q; want the space written only in that one clause", source)
	}
	for _, kept := range []string{"comment_doc_private = params.isPrivate", "attachment_doc_title = params.title", "attachment_doc_reading_users = params.readingUsers"} {
		if !strings.Contains(source, kept) {
			t.Errorf("script %q; want it still to write %s", source, kept)
		}
	}
	if !strings.Contains(query, `"attachment_doc_id": "f1"`) || !strings.Contains(query, `"comment_doc_id": "f1"`) {
		t.Errorf("query %s; want the doc's comments and files", query)
	}
	if _, params, _ := parse(""); params["space"] != "" {
		t.Errorf("space param %v for a doc in no space; want empty, which takes it off", params["space"])
	}
}
