package server

import (
	"os"
	"testing"

	"github.com/google/go-jsonnet"
	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefinition(t *testing.T) {
	testCases := []struct {
		name     string
		params   protocol.DefinitionParams
		expected protocol.DefinitionLink
	}{
		{
			name: "test goto definition for var myvar",
			params: protocol.DefinitionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: "../testdata/test_goto_definition.jsonnet",
					},
					Position: protocol.Position{
						Line:      5,
						Character: 19,
					},
				},
				WorkDoneProgressParams: protocol.WorkDoneProgressParams{},
				PartialResultParams:    protocol.PartialResultParams{},
			},
			expected: protocol.DefinitionLink{
				TargetURI: "../testdata/test_goto_definition.jsonnet",
				TargetRange: protocol.Range{
					Start: protocol.Position{
						Line:      0,
						Character: 6,
					},
					End: protocol.Position{
						Line:      0,
						Character: 15,
					},
				},
				TargetSelectionRange: protocol.Range{
					Start: protocol.Position{
						Line:      0,
						Character: 6,
					},
					End: protocol.Position{
						Line:      0,
						Character: 12,
					},
				},
			},
		},
		{
			name: "test goto definition on function helper",
			params: protocol.DefinitionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: "../testdata/test_goto_definition.jsonnet",
					},
					Position: protocol.Position{
						Line:      7,
						Character: 8,
					},
				},
				WorkDoneProgressParams: protocol.WorkDoneProgressParams{},
				PartialResultParams:    protocol.PartialResultParams{},
			},
			expected: protocol.DefinitionLink{
				TargetURI: "../testdata/test_goto_definition.jsonnet",
				TargetRange: protocol.Range{
					Start: protocol.Position{
						Line:      1,
						Character: 6,
					},
					End: protocol.Position{
						Line:      1,
						Character: 23,
					},
				},
				TargetSelectionRange: protocol.Range{
					Start: protocol.Position{
						Line:      1,
						Character: 6,
					},
					End: protocol.Position{
						Line:      1,
						Character: 13,
					},
				},
			},
		},
		{
			name: "test goto inner definition",
			params: protocol.DefinitionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{
						URI: "../testdata/test_goto_definition_multi_locals.jsonnet",
					},
					Position: protocol.Position{
						Line:      6,
						Character: 11,
					},
				},
				WorkDoneProgressParams: protocol.WorkDoneProgressParams{},
				PartialResultParams:    protocol.PartialResultParams{},
			},
			expected: protocol.DefinitionLink{
				TargetURI: "../testdata/test_goto_definition_multi_locals.jsonnet",
				TargetRange: protocol.Range{
					Start: protocol.Position{
						Line:      4,
						Character: 10,
					},
					End: protocol.Position{
						Line:      4,
						Character: 28,
					},
				},
				TargetSelectionRange: protocol.Range{
					Start: protocol.Position{
						Line:      4,
						Character: 10,
					},
					End: protocol.Position{
						Line:      4,
						Character: 19,
					},
				},
			},
		},
		//{
		//	name: "test goto super",
		//	params: protocol.DefinitionParams{
		//		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
		//			TextDocument: protocol.TextDocumentIdentifier{
		//				URI: "../testdata/test_combined_object.jsonnet",
		//			},
		//			Position: protocol.Position{
		//				Line:      5,
		//				Character: 13,
		//			},
		//		},
		//		WorkDoneProgressParams: protocol.WorkDoneProgressParams{},
		//		PartialResultParams:    protocol.PartialResultParams{},
		//	},
		//	expected: protocol.DefinitionLink{
		//		TargetURI: "../testdata/test_combined_object.jsonnet",
		//		TargetRange: protocol.Range{
		//			Start: protocol.Position{
		//				Line:      1,
		//				Character: 4,
		//			},
		//			End: protocol.Position{
		//				Line:      3,
		//				Character: 5,
		//			},
		//		},
		//		TargetSelectionRange: protocol.Range{
		//			Start: protocol.Position{
		//				Line:      1,
		//				Character: 4,
		//			},
		//			End: protocol.Position{
		//				Line:      3,
		//				Character: 5,
		//			},
		//		},
		//	},
		//},
		//{
		//	name: "test goto super nested",
		//	params: protocol.DefinitionParams{
		//		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
		//			TextDocument: protocol.TextDocumentIdentifier{
		//				URI: "../testdata/test_combined_object.jsonnet",
		//			},
		//			Position: protocol.Position{
		//				Line:      5,
		//				Character: 10,
		//			},
		//		},
		//		WorkDoneProgressParams: protocol.WorkDoneProgressParams{},
		//		PartialResultParams:    protocol.PartialResultParams{},
		//	},
		//	expected: protocol.DefinitionLink{
		//		TargetURI: "../testdata/test_combined_object.jsonnet",
		//		TargetRange: protocol.Range{
		//			Start: protocol.Position{
		//				Line:      2,
		//				Character: 9,
		//			},
		//			End: protocol.Position{
		//				Line:      2,
		//				Character: 23,
		//			},
		//		},
		//		TargetSelectionRange: protocol.Range{
		//			Start: protocol.Position{
		//				Line:      2,
		//				Character: 9,
		//			},
		//			End: protocol.Position{
		//				Line:      2,
		//				Character: 10,
		//			},
		//		},
		//	},
		//},
		//{
		//	name: "test goto self object field function",
		//	params: protocol.DefinitionParams{
		//		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
		//			TextDocument: protocol.TextDocumentIdentifier{
		//				URI: "../testdata/test_basic_lib.libsonnet",
		//			},
		//			Position: protocol.Position{
		//				Line:      4,
		//				Character: 19,
		//			},
		//		},
		//		WorkDoneProgressParams: protocol.WorkDoneProgressParams{},
		//		PartialResultParams:    protocol.PartialResultParams{},
		//	},
		//	expected: protocol.DefinitionLink{
		//		TargetURI: "../testdata/test_basic_lib.libsonnet",
		//		TargetRange: protocol.Range{
		//			Start: protocol.Position{
		//				Line:      1,
		//				Character: 4,
		//			},
		//			End: protocol.Position{
		//				Line:      3,
		//				Character: 20,
		//			},
		//		},
		//		TargetSelectionRange: protocol.Range{
		//			Start: protocol.Position{
		//				Line:      1,
		//				Character: 4,
		//			},
		//			End: protocol.Position{
		//				Line:      1,
		//				Character: 10,
		//			},
		//		},
		//	},
		//},
		//{
		//	name: "test goto super object field ",
		//	params: protocol.DefinitionParams{
		//		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
		//			TextDocument: protocol.TextDocumentIdentifier{
		//				URI: "./testdata/oo-contrived.jsonnet",
		//			},
		//			Position: protocol.Position{
		//				Line:      12,
		//				Character: 17,
		//			},
		//		},
		//		WorkDoneProgressParams: protocol.WorkDoneProgressParams{},
		//		PartialResultParams:    protocol.PartialResultParams{},
		//	},
		//	expected: protocol.DefinitionLink{
		//		TargetURI: "./testdata/oo-contrived.jsonnet",
		//		TargetRange: protocol.Range{
		//			Start: protocol.Position{
		//				Line:      1,
		//				Character: 2,
		//			},
		//			End: protocol.Position{
		//				Line:      1,
		//				Character: 6,
		//			},
		//		},
		//		TargetSelectionRange: protocol.Range{
		//			Start: protocol.Position{
		//				Line:      1,
		//				Character: 2,
		//			},
		//			End: protocol.Position{
		//				Line:      1,
		//				Character: 3,
		//			},
		//		},
		//	},
		//},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filename := string(tc.params.TextDocument.URI)
			var content, err = os.ReadFile(filename)
			require.NoError(t, err)

			ast, err := jsonnet.SnippetToAST(filename, string(content))
			require.NoError(t, err)
			got, err := Definition(ast, tc.params)
			assert.Equal(t, tc.expected, got)
		})
	}
}
