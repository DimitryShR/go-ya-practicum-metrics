package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/imports"
)

type structInfo struct {
	pkgName string
	name    string
	fields  []fieldInfo
}

type fieldInfo struct {
	name     string // empty for embedded fields
	expr     ast.Expr
	embedded bool
}

// известные примитивные типы (built-in)
var builtinTypes = map[string]string{
	"int":     "0",
	"int8":    "0",
	"int16":   "0",
	"int32":   "0",
	"int64":   "0",
	"uint":    "0",
	"uint8":   "0",
	"uint16":  "0",
	"uint32":  "0",
	"uint64":  "0",
	"float32": "0",
	"float64": "0",
	"string":  `""`,
	"bool":    "false",
	"byte":    "0",
	"rune":    "0",
	"uintptr": "0",
}

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	// сканируем все директории и собираем структуры
	pkgStructs := make(map[string][]structInfo) // key: directory path

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		// пропускаем скрытые директории
		if strings.HasPrefix(d.Name(), ".") && path != "." {
			return filepath.SkipDir
		}
		// пропускаем vendor
		if d.Name() == "vendor" {
			return filepath.SkipDir
		}

		// ищем .go файлы в директории
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		var goFiles []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") && e.Name() != "reset.gen.go" {
				goFiles = append(goFiles, filepath.Join(path, e.Name()))
			}
		}
		if len(goFiles) == 0 {
			return nil
		}

		structs, err := parseStructs(goFiles)
		if err != nil {
			return fmt.Errorf("error parsing %s: %w", path, err)
		}
		if len(structs) > 0 {
			pkgStructs[path] = structs
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// генерируем файлы для каждого пакета
	for dir, structs := range pkgStructs {
		// собираем все структуры пакета для определения вложенных структур
		pkgTypes := collectTypesFromDir(dir)
		outPath := filepath.Join(dir, "reset.gen.go")
		content, err := generateFile(structs[0].pkgName, structs, pkgTypes, outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating code for %s: %v\n", dir, err)
			os.Exit(1)
		}

		if err := os.WriteFile(outPath, content, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outPath, err)
			os.Exit(1)
		}
		fmt.Printf("Generated: %s\n", outPath)
	}
}

// parseStructs парсит список .go файлов и возвращает структуры с // generate:reset
func parseStructs(files []string) ([]structInfo, error) {
	var structs []structInfo
	fset := token.NewFileSet()

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		for _, d := range f.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}

			// проверяем наличие // generate:reset в блоке объявления типа
			if !hasGenerateReset(gd.Doc) {
				continue
			}

			for _, s := range gd.Specs {
				ts, ok := s.(*ast.TypeSpec)
				if !ok {
					continue
				}

				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}

				fields := parseFields(st.Fields)

				structs = append(structs, structInfo{
					pkgName: f.Name.Name,
					name:    ts.Name.Name,
					fields:  fields,
				})
			}
		}
	}

	return structs, nil
}

// hasGenerateReset проверяет, содержит ли группа комментарий // generate:reset
func hasGenerateReset(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	for _, comment := range cg.List {
		text := strings.TrimSpace(comment.Text)
		// Проверяем // generate:reset или //generate:reset args
		if text == "// generate:reset" || strings.HasPrefix(text, "// generate:reset ") {
			return true
		}
	}
	return false
}

// parseFields извлекает информацию о полях структуры
func parseFields(fieldList *ast.FieldList) []fieldInfo {
	if fieldList == nil {
		return nil
	}

	var fields []fieldInfo
	for _, f := range fieldList.List {
		if len(f.Names) == 0 {
			// встроенное (embedded) поле
			fields = append(fields, fieldInfo{
				name:     "",
				expr:     f.Type,
				embedded: true,
			})
		} else {
			for _, n := range f.Names {
				fields = append(fields, fieldInfo{
					name:     n.Name,
					expr:     f.Type,
					embedded: false,
				})
			}
		}
	}
	return fields
}

// collectTypesFromDir собирает информацию о типах в директории
func collectTypesFromDir(dir string) map[string]ast.Expr {
	types := make(map[string]ast.Expr)
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.IsDir() ||
			!strings.HasSuffix(e.Name(), ".go") ||
			strings.HasSuffix(e.Name(), "_test.go") ||
			e.Name() == "reset.gen.go" {
			continue
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, filepath.Join(dir, e.Name()), nil, parser.ParseComments)
		if err != nil {
			continue
		}
		for _, d := range f.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, s := range gd.Specs {
				ts, ok := s.(*ast.TypeSpec)
				if ok {
					types[ts.Name.Name] = ts.Type
				}
			}
		}
	}
	return types
}

// generateFile генерирует содержимое reset.gen.go для пакета
func generateFile(pkgName string, structs []structInfo, pkgTypes map[string]ast.Expr, filename string) ([]byte, error) {
	var buf bytes.Buffer

	// пишем заголовок
	buf.WriteString("// Code generated by go generate; DO NOT EDIT.\n")
	buf.WriteString("// This file was generated by cmd/reset/main.go\n\n")
	fmt.Fprintf(&buf, "package %s\n\n", pkgName)

	// сортируем структуры для стабильности генерации
	sort.Slice(structs, func(i, j int) bool {
		return structs[i].name < structs[j].name
	})

	for _, st := range structs {
		buf.WriteString(generateResetMethod(st, structs, pkgTypes))
		buf.WriteString("\n")
	}

	// Обрабатываем через imports (форматирование + добавление/удаление импортов)
	formatted, err := imports.Process(filename, buf.Bytes(), nil)
	if err != nil {
		return nil, fmt.Errorf("imports.Process error: %w\n%s", err, buf.String())
	}

	return formatted, nil
}

// generateResetMethod генерирует метод Reset() для одной структуры
func generateResetMethod(st structInfo, allStructs []structInfo, pkgTypes map[string]ast.Expr) string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "func (s *%s) Reset() {\n", st.name)
	buf.WriteString("\tif s == nil {\n")
	buf.WriteString("\t\treturn\n")
	buf.WriteString("\t}\n\n")

	for _, f := range st.fields {
		buf.WriteString(generateFieldReset(f, "s", allStructs, pkgTypes))
	}

	buf.WriteString("}\n")
	return buf.String()
}

// generateFieldReset генерирует код сброса для одного поля
func generateFieldReset(f fieldInfo, prefix string, allStructs []structInfo, pkgTypes map[string]ast.Expr) string {
	var fieldRef string
	if f.embedded {
		fieldName := embeddedFieldName(f.expr)
		if fieldName == "" {
			// fallback — не удалось извлечь имя встроенного поля
			return fmt.Sprintf("\t// WARNING: embedded field name could not be resolved for type %s\n",
				exprToString(f.expr))
		}
		fieldRef = prefix + "." + fieldName
	} else {
		fieldRef = prefix + "." + f.name
	}

	return renderFieldExpr(f.expr, fieldRef, allStructs, pkgTypes)
}

// embeddedFieldName возвращает неявное имя встроенного поля по его типу.
// Например: Bar → "Bar", *Bar → "Bar", time.Time → "Time", pkg.Bar → "Bar".
func embeddedFieldName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return e.Sel.Name
	case *ast.StarExpr:
		return embeddedFieldName(e.X)
	default:
		return ""
	}
}

// renderFieldExpr генерирует код сброса для выражения типа
func renderFieldExpr(expr ast.Expr, fieldRef string, allStructs []structInfo, pkgTypes map[string]ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		// примитив или тип из того же пакета
		if zeroVal, ok := builtinTypes[t.Name]; ok {
			return fmt.Sprintf("\t%s = %s\n", fieldRef, zeroVal)
		}
		// возможно это тип (псевдоним) из того же пакета
		if typeExpr, ok := pkgTypes[t.Name]; ok {
			switch underlying := typeExpr.(type) {
			case *ast.StructType:
				// структура
				if structHasReset(allStructs, t.Name) {
					return fmt.Sprintf("\t%s.Reset()\n", fieldRef)
				}
				// рекурсивно разворачиваем вложенную структуру (без Reset)
				return generateInlineReset(underlying, fieldRef, allStructs, pkgTypes)
			case *ast.Ident:
				// псевдоним примитива (type MetricType string)
				if zeroVal, ok := builtinTypes[underlying.Name]; ok {
					return fmt.Sprintf("\t%s = %s\n", fieldRef, zeroVal)
				}
				// псевдоним другого идентификатора из того же пакета
				return fmt.Sprintf("\t%s = *new(%s)\n", fieldRef, t.Name)
			case *ast.SelectorExpr:
				// псевдоним внешнего типа
				return fmt.Sprintf("\t%s = *new(%s)\n", fieldRef, exprToString(typeExpr))
			}
		}
		// неизвестный тип — сбрасываем к нулевому значению
		return fmt.Sprintf("\t%s = *new(%s)\n", fieldRef, t.Name)

	case *ast.StarExpr:
		// указатель
		var resetCode string
		switch inner := t.X.(type) {
		case *ast.Ident:
			if _, ok := builtinTypes[inner.Name]; ok {
				zeroVal := builtinTypes[inner.Name]
				resetCode = fmt.Sprintf("\t\t*%s = %s\n", fieldRef, zeroVal)
			} else if typeExpr, ok := pkgTypes[inner.Name]; ok {
				switch underlying := typeExpr.(type) {
				case *ast.StructType:
					if structHasReset(allStructs, inner.Name) {
						resetCode = fmt.Sprintf("\t\t%s.Reset()\n", fieldRef)
					} else {
						// структура без Reset — обнуляем через *new
						resetCode = fmt.Sprintf("\t\t*%s = *new(%s)\n", fieldRef, inner.Name)
					}
				case *ast.Ident:
					if zeroVal, ok := builtinTypes[underlying.Name]; ok {
						resetCode = fmt.Sprintf("\t\t*%s = %s\n", fieldRef, zeroVal)
					} else {
						resetCode = fmt.Sprintf("\t\t*%s = *new(%s)\n", fieldRef, inner.Name)
					}
				default:
					resetCode = fmt.Sprintf("\t\t*%s = *new(%s)\n", fieldRef, inner.Name)
				}
			} else {
				resetCode = fmt.Sprintf("\t\t*%s = *new(%s)\n", fieldRef, inner.Name)
			}
		case *ast.ArrayType:
			resetCode = fmt.Sprintf("\t\t*%s = (*%s)[:0]\n", fieldRef, fieldRef)
		case *ast.MapType:
			resetCode = fmt.Sprintf("\t\tclear(*%s)\n", fieldRef)
		case *ast.SelectorExpr:
			resetCode = fmt.Sprintf("\t\t*%s = *new(%s)\n", fieldRef, exprToString(inner))
		case *ast.StarExpr:
			resetCode = fmt.Sprintf("\t\t*%s = *new(%s)\n", fieldRef, exprToString(inner))
		default:
			resetCode = fmt.Sprintf("\t\t*%s = *new(%s)\n", fieldRef, exprToString(inner))
		}
		return fmt.Sprintf("\tif %s != nil {\n%s\t}\n", fieldRef, resetCode)

	case *ast.ArrayType:
		// срез
		return fmt.Sprintf("\t%s = %s[:0]\n", fieldRef, fieldRef)

	case *ast.MapType:
		// мапа
		return fmt.Sprintf("\tclear(%s)\n", fieldRef)

	case *ast.SelectorExpr:
		// тип из другого пакета (например time.Duration)
		return fmt.Sprintf("\t%s = *new(%s)\n", fieldRef, exprToString(t))

	}

	// fallback: ни одна ветка не сработала — универсальное обнуление через *new
	return fmt.Sprintf("\t// WARNING: unknown field type, using zero value\n\t%s = *new(%s)\n",
		fieldRef, exprToString(expr))

}

// structHasReset проверяет, есть ли в списке структур структура с методом Reset
// (т.е. помечена // generate:reset)
func structHasReset(structs []structInfo, typeName string) bool {
	for _, s := range structs {
		if s.name == typeName {
			return true
		}
	}
	return false
}

// exprToString преобразует ast.Expr в строковое представление типа.
func exprToString(expr ast.Expr) string {
	var buf strings.Builder
	if err := printer.Fprint(&buf, token.NewFileSet(), expr); err == nil {
		return buf.String()
	}
	return "interface{}"
}

// generateInlineReset генерирует код для рекурсивного сброса вложенной структуры без Reset()
func generateInlineReset(expr ast.Expr, fieldRef string, allStructs []structInfo, pkgTypes map[string]ast.Expr) string {
	st, ok := expr.(*ast.StructType)
	if !ok {
		return ""
	}
	if st.Fields == nil {
		return ""
	}

	var buf bytes.Buffer

	for _, f := range st.Fields.List {
		var names []string
		if len(f.Names) == 0 {
			// embedded поле
			names = []string{""}
		} else {
			for _, n := range f.Names {
				names = append(names, n.Name)
			}
		}

		for _, name := range names {
			var innerFieldRef string
			if name == "" {
				// Встроенное поле – используем неявное имя типа
				fieldName := embeddedFieldName(f.Type)
				if fieldName == "" {
					// fallback — не удалось извлечь имя встроенного поля
					buf.WriteString(
						fmt.Sprintf("\t// WARNING: embedded field name could not be resolved for type %s inside %s\n",
							exprToString(f.Type),
							fieldRef))
					continue
				}
				innerFieldRef = fieldRef + "." + fieldName
			} else {
				innerFieldRef = fieldRef + "." + name
			}

			// Генерируем код сброса для данного поля
			buf.WriteString(renderFieldExpr(f.Type, innerFieldRef, allStructs, pkgTypes))
		}
	}

	return buf.String()
}
