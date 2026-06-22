package pkgmanager

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

type PackageLoader struct {
	SavePath   string
	LocalCache string
}

func NewPackageLoader(savePath, localCache string) *PackageLoader {
	return &PackageLoader{
		SavePath:   savePath,
		LocalCache: localCache,
	}
}

// cleanTypeString убирает префиксы "untyped" и полное имя пакета.
func cleanTypeString(t string) string {
	t = strings.TrimPrefix(t, "untyped ")
	if idx := strings.LastIndex(t, "."); idx != -1 {
		return t[idx+1:]
	}
	return t
}

// packageData хранит собранные экспортируемые элементы пакета.
type packageData struct {
	Funcs     []string
	Constants []string
	Structs   []string
	Variables []string
}

// Load загружает и анализирует пакет, сохраняя результаты в файлы.
func (pl *PackageLoader) Load(packageName string) error {
	if pl.isPackageCached(packageName) {
		return nil
	}

	absPath, err := pl.resolvePackagePath(packageName)
	if err != nil {
		return err
	}

	data, err := pl.loadAndAnalyzePackage(absPath)
	if err != nil {
		return err
	}

	return pl.savePackageData(packageName, data)
}

// isPackageCached проверяет наличие файла func.txt в кеше.
func (pl *PackageLoader) isPackageCached(packageName string) bool {
	path := filepath.Join(pl.SavePath, packageName)
	_, err := os.Stat(filepath.Join(path, "func.txt"))
	return err == nil
}

// resolvePackagePath получает абсолютный путь к пакету.
func (pl *PackageLoader) resolvePackagePath(packageName string) (string, error) {
	fmt.Printf("Resolving path for %s...\n", packageName)
	absPath, err := ResolvePackagePath(packageName, pl.LocalCache)
	if err != nil {
		return "", fmt.Errorf("resolution failed: %w", err)
	}
	fmt.Printf("Analyzing package at: %s\n", absPath)
	return absPath, nil
}

// loadAndAnalyzePackage загружает пакет через go/packages и анализирует его.
func (pl *PackageLoader) loadAndAnalyzePackage(absPath string) (*packageData, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedSyntax,
		Dir: absPath,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to load package from %s: %w", absPath, err)
	}

	fmt.Printf("Found %d package(s)\n", len(pkgs))
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found at %s", absPath)
	}

	pkg := pkgs[0]
	fmt.Printf("Analyzing package: %s\n", pkg.Name)
	fmt.Printf("Loaded syntax for %d files\n", len(pkg.Syntax))

	return pl.analyzePackage(pkg), nil
}

// analyzePackage обходит все файлы пакета и собирает экспортируемые элементы.
func (pl *PackageLoader) analyzePackage(pkg *packages.Package) *packageData {
	data := &packageData{}
	for _, file := range pkg.Syntax {
		pl.analyzeFile(file, pkg, data)
	}
	return data
}

// analyzeFile обходит AST одного файла и заполняет data.
func (pl *PackageLoader) analyzeFile(file *ast.File, pkg *packages.Package, data *packageData) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			pl.handleGenDecl(x, pkg, data)
		case *ast.FuncDecl:
			pl.handleFuncDecl(x, pkg, data)
		}
		return true
	})
}

// handleGenDecl обрабатывает общие объявления (const, var, type).
func (pl *PackageLoader) handleGenDecl(decl *ast.GenDecl, pkg *packages.Package, data *packageData) {
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.ValueSpec:
			pl.handleValueSpec(s, decl.Tok, pkg, data)
		case *ast.TypeSpec:
			pl.handleTypeSpec(s, pkg, data)
		}
	}
}

// handleValueSpec обрабатывает спецификации значений (константы и переменные).
func (pl *PackageLoader) handleValueSpec(spec *ast.ValueSpec, tok token.Token, pkg *packages.Package, data *packageData) {
	for _, name := range spec.Names {
		if !name.IsExported() {
			continue
		}
		obj := pkg.TypesInfo.Defs[name]
		typeStr := "unknown"
		if obj != nil && obj.Type() != nil {
			typeStr = cleanTypeString(obj.Type().String())
		}

		entry := fmt.Sprintf("%s %s", name.Name, typeStr)
		switch tok { //nolint: exhaustive
		case token.CONST:
			data.Constants = append(data.Constants, entry)
		case token.VAR:
			data.Variables = append(data.Variables, entry)
		}
	}
}

// typeToString рекурсивно преобразует AST выражение типа в строку.
func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		prefix := ""
		if t.Len == nil {
			prefix = "[]"
		} else {
			// Для упрощения выводим размер, если он есть
			prefix = fmt.Sprintf("[%d]", 0) // В реальности нужно обработать t.Len (ast.Expr)
		}
		return prefix + typeToString(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", typeToString(t.Key), typeToString(t.Value))
	case *ast.SelectorExpr:
		// Для селекторов возвращаем только имя самого типа (аналог cleanTypeString)
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
		return t.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	default:
		return "unknown"
	}
}

// handleTypeSpec обрабатывает определения типов (структуры).
func (pl *PackageLoader) handleTypeSpec(spec *ast.TypeSpec, pkg *packages.Package, data *packageData) {
	if !spec.Name.IsExported() {
		return
	}
	stTyp, ok := spec.Type.(*ast.StructType)
	if !ok {
		return
	}

	var fields []string
	for _, field := range stTyp.Fields.List {
		typeStr := typeToString(field.Type)
		// Если это именованный тип из другого пакета, он может прийти как SelectorExpr.
		// В данном упрощенном AST подходе мы просто берем имя.
		if idx := strings.LastIndex(typeStr, "."); idx != -1 {
			typeStr = typeStr[idx+1:]
		}

		for _, name := range field.Names {
			if name.IsExported() {
				fields = append(fields, fmt.Sprintf("%s %s", name.Name, typeStr))
			}
		}
	}

	structDef := fmt.Sprintf("%s {\n  %s\n}", spec.Name.Name, strings.Join(fields, "\n  "))
	data.Structs = append(data.Structs, structDef)
}

// handleFuncDecl обрабатывает объявления функций (включая методы).
func (pl *PackageLoader) handleFuncDecl(decl *ast.FuncDecl, pkg *packages.Package, data *packageData) {
	if !decl.Name.IsExported() {
		return
	}

	obj := pkg.TypesInfo.Defs[decl.Name]
	if obj == nil {
		return
	}
	fn, ok := obj.(*types.Func)
	if !ok {
		return
	}

	sig := fn.Type().String()
	sig = strings.ReplaceAll(sig, pkg.PkgPath+".", "")
	prefix := pl.extractReceiverPrefix(decl)

	entry := fmt.Sprintf("%s%s: %s", prefix, decl.Name.Name, sig)
	data.Funcs = append(data.Funcs, entry)
}

// extractReceiverPrefix извлекает имя получателя для метода.
func (pl *PackageLoader) extractReceiverPrefix(decl *ast.FuncDecl) string {
	if decl.Recv == nil || len(decl.Recv.List) == 0 {
		return ""
	}
	recvExpr := decl.Recv.List[0].Type
	var recvName string
	switch rt := recvExpr.(type) {
	case *ast.StarExpr:
		if it, ok := rt.X.(*ast.Ident); ok {
			recvName = it.Name
		}
	case *ast.Ident:
		recvName = rt.Name
	}
	if recvName != "" {
		return recvName + "."
	}
	return ""
}

// savePackageData сохраняет собранные данные в файлы.
func (pl *PackageLoader) savePackageData(packageName string, data *packageData) error {
	if err := pl.saveToFile(packageName, "func.txt", strings.Join(data.Funcs, "\n")); err != nil {
		return err
	}
	if err := pl.saveToFile(packageName, "const.txt", strings.Join(data.Constants, "\n")); err != nil {
		return err
	}
	if err := pl.saveToFile(packageName, "struct.txt", strings.Join(data.Structs, "\n")); err != nil {
		return err
	}
	if err := pl.saveToFile(packageName, "var.txt", strings.Join(data.Variables, "\n")); err != nil {
		return err
	}
	return nil
}

// saveToFile записывает содержимое в файл внутри директории пакета.
func (pl *PackageLoader) saveToFile(packageName, filename, content string) error {
	const dirPerms = 0o755
	const filePerms = 0o600

	dir := filepath.Join(pl.SavePath, packageName)
	if err := os.MkdirAll(dir, dirPerms); err != nil {
		return fmt.Errorf("failed to create directory for package %s: %v", packageName, err)
	}

	fullPath := filepath.Join(dir, filename)
	return os.WriteFile(fullPath, []byte(content), filePerms)
}
