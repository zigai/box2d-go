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
	URL:      "https://github.com/zigai/box2d/tree/master/internal/lint",
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
			if (n.Tok != token.ADD_ASSIGN && n.Tok != token.SUB_ASSIGN) || len(n.Lhs) != 1 || len(n.Rhs) != 1 {
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
	typ := types.TypeString(pass.TypesInfo.TypeOf(mul), types.RelativeTo(pass.Pkg))
	pass.Report(analysis.Diagnostic{
		Pos:     mul.Pos(),
		End:     mul.End(),
		Message: "float32 product feeds " + consumer + " without explicit rounding; wrap it in " + typ + "(...) to prevent fused multiply-add",
		SuggestedFixes: []analysis.SuggestedFix{{
			Message: "Round the product with " + typ + "(...)",
			TextEdits: []analysis.TextEdit{
				{Pos: mul.Pos(), End: mul.Pos(), NewText: []byte(typ + "(")},
				{Pos: mul.End(), End: mul.End(), NewText: []byte(")")},
			},
		}},
	})
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
