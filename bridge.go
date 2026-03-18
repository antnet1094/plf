package main

import "C"
import (
	"encoding/json"
	"github.com/antnet1094/plf/pkg/parser"
	"github.com/antnet1094/plf/pkg/renderer"
	"github.com/antnet1094/plf/pkg/types"
)

// RenderPLF exporta la funcionalidad de renderizado a C/Python.
// Recibe la ruta del archivo y un JSON con las variables.
// Devuelve el prompt final renderizado como una cadena de C.
//export RenderPLF
func RenderPLF(plfPath *C.char, varsJson *C.char) *C.char {
	path := C.GoString(plfPath)
	varsStr := C.GoString(varsJson)

	// 1. Parsear el archivo PLF
	doc, err := parser.ParseFile(path)
	if err != nil {
		return C.CString("ERROR_PARSE: " + err.Error())
	}

	// 2. Decodificar las variables desde el JSON de Python
	var vars map[string]string
	if err := json.Unmarshal([]byte(varsStr), &vars); err != nil {
		return C.CString("ERROR_VARS_JSON: " + err.Error())
	}

	// 3. Renderizar el prompt (usamos formato RAW por defecto para Python)
	result, err := renderer.Render(doc, types.RenderOptions{
		Vars:   vars,
		Format: types.FormatRaw,
		Minify: true, // Minificamos para ahorrar tokens
	})
	if err != nil {
		return C.CString("ERROR_RENDER: " + err.Error())
	}

	// 4. Devolver el resultado como cadena de C
	return C.CString(result.Full)
}

func main() {} // Requerido para construir bibliotecas compartidas
