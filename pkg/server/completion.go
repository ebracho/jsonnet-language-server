package server

import (
	"context"
	"reflect"
	"sort"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/grafana/jsonnet-language-server/pkg/ast/processing"
	"github.com/grafana/jsonnet-language-server/pkg/nodestack"
	position "github.com/grafana/jsonnet-language-server/pkg/position_conversion"
	"github.com/grafana/jsonnet-language-server/pkg/utils"
	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	log "github.com/sirupsen/logrus"
)

func (s *Server) Completion(_ context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	doc, err := s.cache.get(params.TextDocument.URI)
	if err != nil {
		return nil, utils.LogErrorf("Completion: %s: %w", errorRetrievingDocument, err)
	}

	line := getCompletionLine(doc.item.Text, params.Position)

	// Short-circuit if it's a stdlib completion
	if items := s.completionStdLib(line); len(items) > 0 {
		return &protocol.CompletionList{IsIncomplete: false, Items: items}, nil
	}

	// Otherwise, parse the AST and search for completions
	if doc.ast == nil {
		log.Errorf("Completion: document was never successfully parsed, can't autocomplete")
		return nil, nil
	}

	searchStack, err := processing.FindNodeByPosition(doc.ast, position.ProtocolToAST(params.Position))
	if err != nil {
		log.Errorf("Completion: error computing node: %v", err)
		return nil, nil
	}

	vm := s.getVM(doc.item.URI.SpanURI().Filename())

	items := s.completionFromStack(line, searchStack, vm)
	return &protocol.CompletionList{IsIncomplete: false, Items: items}, nil
}

func getCompletionLine(fileContent string, position protocol.Position) string {
	line := strings.Split(fileContent, "\n")[position.Line]
	charIndex := int(position.Character)
	if charIndex > len(line) {
		charIndex = len(line)
	}
	line = line[:charIndex]
	return line
}

func (s *Server) completionFromStack(line string, stack *nodestack.NodeStack, vm *jsonnet.VM) []protocol.CompletionItem {
	lineWords := strings.Split(line, " ")
	lastWord := lineWords[len(lineWords)-1]
	lastWord = strings.TrimRight(lastWord, ",;") // Ignore trailing commas and semicolons, they can present when someone is modifying an existing line

	indexes := strings.Split(lastWord, ".")

	if len(indexes) == 1 {
		var items []protocol.CompletionItem
		// firstIndex is a variable (local) completion
		for !stack.IsEmpty() {
			if curr, ok := stack.Pop().(*ast.Local); ok {
				for _, bind := range curr.Binds {
					label := string(bind.Variable)

					if !strings.HasPrefix(label, indexes[0]) {
						continue
					}

					items = append(items, createCompletionItem(label, label, protocol.VariableCompletion, bind.Body))
				}
			}
		}
		return items
	}

	ranges, err := processing.FindRangesFromIndexList(stack, indexes, vm, true)
	if err != nil {
		log.Errorf("Completion: error finding ranges: %v", err)
		return nil
	}

	completionPrefix := strings.Join(indexes[:len(indexes)-1], ".")
	return createCompletionItemsFromRanges(ranges, completionPrefix, line)
}

func (s *Server) completionStdLib(line string) []protocol.CompletionItem {
	var items []protocol.CompletionItem

	stdIndex := strings.LastIndex(line, "std.")
	if stdIndex != -1 {
		userInput := line[stdIndex+4:]
		funcStartWith := []protocol.CompletionItem{}
		funcContains := []protocol.CompletionItem{}
		for _, f := range s.stdlib {
			if f.Name == userInput {
				break
			}
			lowerFuncName := strings.ToLower(f.Name)
			findName := strings.ToLower(userInput)
			item := protocol.CompletionItem{
				Label:         f.Name,
				Kind:          protocol.FunctionCompletion,
				Detail:        f.Signature(),
				InsertText:    strings.ReplaceAll(f.Signature(), "std.", ""),
				Documentation: f.MarkdownDescription,
			}

			if len(findName) > 0 && strings.HasPrefix(lowerFuncName, findName) {
				funcStartWith = append(funcStartWith, item)
				continue
			}

			if strings.Contains(lowerFuncName, findName) {
				funcContains = append(funcContains, item)
			}
		}

		items = append(items, funcStartWith...)
		items = append(items, funcContains...)
	}

	return items
}

func createCompletionItemsFromRanges(ranges []processing.ObjectRange, completionPrefix, currentLine string) []protocol.CompletionItem {
	var items []protocol.CompletionItem
	labels := make(map[string]bool)

	for _, field := range ranges {
		label := field.FieldName

		if field.Node == nil {
			continue
		}

		if labels[label] {
			continue
		}

		// Ignore the current field
		if strings.Contains(currentLine, label+":") && completionPrefix == "self" {
			continue
		}

		items = append(items, createCompletionItem(label, completionPrefix+"."+label, protocol.FieldCompletion, field.Node))
		labels[label] = true
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Label < items[j].Label
	})

	return items
}

func createCompletionItem(label, detail string, kind protocol.CompletionItemKind, body ast.Node) protocol.CompletionItem {
	insertText := label
	if asFunc, ok := body.(*ast.Function); ok {
		kind = protocol.FunctionCompletion
		params := []string{}
		for _, param := range asFunc.Parameters {
			params = append(params, string(param.Name))
		}
		paramsString := "(" + strings.Join(params, ", ") + ")"
		detail += paramsString
		insertText += paramsString
	}

	return protocol.CompletionItem{
		Label:  label,
		Detail: detail,
		Kind:   kind,
		LabelDetails: protocol.CompletionItemLabelDetails{
			Description: typeToString(body),
		},
		InsertText: insertText,
	}
}

func typeToString(t ast.Node) string {
	switch t.(type) {
	case *ast.Array:
		return "array"
	case *ast.LiteralBoolean:
		return "boolean"
	case *ast.Function:
		return "function"
	case *ast.LiteralNull:
		return "null"
	case *ast.LiteralNumber:
		return "number"
	case *ast.Object, *ast.DesugaredObject:
		return "object"
	case *ast.LiteralString:
		return "string"
	case *ast.Import, *ast.ImportStr:
		return "import"
	}
	return reflect.TypeOf(t).String()
}
