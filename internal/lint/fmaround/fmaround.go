package fmaround

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "fmaround",
	Doc:      "flag float32 products that feed an addition or subtraction without explicit float32(...) rounding",
	URL:      "https://github.com/zigai/box2d-go/tree/master/internal/lint",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	in := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	in.Preorder([]ast.Node{(*ast.BinaryExpr)(nil), (*ast.AssignStmt)(nil)}, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.BinaryExpr:
			if n.Op != token.ADD && n.Op != token.SUB {
				return
			}
			if !isFloat32(pass.TypesInfo.TypeOf(n)) {
				return
			}
			report(pass, n.X, "an addition or subtraction")
			report(pass, n.Y, "an addition or subtraction")
		case *ast.AssignStmt:
			if (n.Tok != token.ADD_ASSIGN && n.Tok != token.SUB_ASSIGN) || len(n.Lhs) != 1 ||
				len(n.Rhs) != 1 {
				return
			}
			if !isFloat32(pass.TypesInfo.TypeOf(n.Lhs[0])) {
				return
			}
			report(pass, n.Rhs[0], n.Tok.String())
		}
	})
	return nil, nil
}

func report(pass *analysis.Pass, operand ast.Expr, consumer string) {
	mul := product(pass.TypesInfo, operand)
	if mul == nil {
		return
	}
	typ := roundingType(pass, mul)
	diagnostic := analysis.Diagnostic{
		Pos:     mul.Pos(),
		End:     mul.End(),
		Message: "float32 product feeds " + consumer + " without explicit rounding",
	}
	if typ != "" {
		diagnostic.Message += "; wrap it in " + typ + "(...) to prevent fused multiply-add"
		diagnostic.SuggestedFixes = []analysis.SuggestedFix{{
			Message: "Round the product with " + typ + "(...)",
			TextEdits: []analysis.TextEdit{
				{Pos: mul.Pos(), End: mul.Pos(), NewText: []byte(typ + "(")},
				{Pos: mul.End(), End: mul.End(), NewText: []byte(")")},
			},
		}}
	}
	pass.Report(diagnostic)
}

// A type's declared package name is not necessarily its spelling at the edit
// site: imports can be renamed or dotted, and type names can be shadowed.
func roundingType(pass *analysis.Pass, product ast.Expr) string {
	target := pass.TypesInfo.TypeOf(product)
	scope := pass.Pkg.Scope().Innermost(product.Pos())
	if scope == nil {
		return ""
	}
	qualifier := func(pkg *types.Package) string {
		if pkg.Path() == pass.Pkg.Path() {
			return ""
		}
		for s := scope; s != nil; s = s.Parent() {
			for _, name := range s.Names() {
				obj, ok := s.Lookup(name).(*types.PkgName)
				if !ok || obj.Imported().Path() != pkg.Path() {
					continue
				}
				if _, visible := scope.LookupParent(name, product.Pos()); visible == obj {
					return name
				}
			}
		}
		return "" // A dot-imported type is unqualified; Eval checks visibility.
	}
	name := types.TypeString(target, qualifier)
	if value, err := types.Eval(pass.Fset, pass.Pkg, product.Pos(), name); err == nil &&
		value.IsType() && types.Identical(value.Type, target) {
		return name
	}
	for s := scope; s != nil; s = s.Parent() {
		for _, name := range s.Names() {
			obj, ok := s.Lookup(name).(*types.TypeName)
			if !ok || !types.Identical(obj.Type(), target) {
				continue
			}
			if _, visible := scope.LookupParent(name, product.Pos()); visible == obj {
				return name
			}
		}
	}
	return "" // Report the defect without offering a conversion that cannot compile.
}

func product(info *types.Info, e ast.Expr) *ast.BinaryExpr {
	for {
		switch x := e.(type) {
		case *ast.ParenExpr:
			e = x.X
		case *ast.UnaryExpr:
			if x.Op != token.SUB && x.Op != token.ADD {
				return nil
			}
			e = x.X
		case *ast.BinaryExpr:
			if x.Op != token.MUL || !isFloat32(info.TypeOf(x)) {
				return nil
			}
			if tv, ok := info.Types[x]; ok && tv.Value != nil {
				return nil
			}
			return x
		default:
			return nil
		}
	}
}

func isFloat32(t types.Type) bool {
	if t == nil {
		return false
	}
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Kind() == types.Float32
}
