package types

import (
	"github.com/vs-ude/tinyfyr/internal/errlog"
	"github.com/vs-ude/tinyfyr/internal/lexer"
	"github.com/vs-ude/tinyfyr/internal/parser"
)

func declareGenericFunction(ast *parser.FuncNode, s *Scope, log *errlog.ErrorLog) (*GenericFunc, error) {
	if ast.GenericParams == nil {
		panic("Wrong")
	}
	var cmp *ComponentType
	if cmpScope := s.ComponentScope(); cmpScope != nil {
		cmp = cmpScope.Component
	}
	f := &GenericFunc{name: ast.NameToken.StringValue, ast: ast, Component: cmp}
	for _, p := range ast.GenericParams.Params {
		f.TypeParameters = append(f.TypeParameters, &GenericTypeParameter{Name: p.NameToken.StringValue})
	}
	if err := parseGenericFuncAttribs(ast, f, log); err != nil {
		return nil, err
	}
	return f, s.AddElement(f, ast.Location(), log)
}

func declareFunction(ast *parser.FuncNode, s *Scope, log *errlog.ErrorLog) (*Func, error) {
	if ast.GenericParams != nil {
		panic("Wrong")
	}
	var err error
	var name string
	// Destructor ?
	if ast.TildeToken != nil {
		name = "__dtor__"
	} else {
		name = ast.NameToken.StringValue
	}
	loc := ast.Location()
	ft := &FuncType{TypeBase: TypeBase{name: name, location: loc, pkg: s.PackageScope().Package}}
	// Destructor ?
	if ast.TildeToken != nil {
		ft.IsDestructor = true
	}
	f := &Func{name: name, Type: ft, Ast: ast, OuterScope: s, Location: loc, Component: s.Component}
	f.InnerScope = newScope(f.OuterScope, FunctionScope, f.Location)
	f.InnerScope.Func = f
	if ast.Type != nil {
		if mt, ok := ast.Type.(*parser.MutableTypeNode); ok && mt.MutToken.Kind == lexer.TokenDual {
			if s.dualIsMut != -1 {
				f.DualIsMut = true
				s.dualIsMut = 1
				f.InnerScope.dualIsMut = 1
			} else {
				f.InnerScope.dualIsMut = -1
			}
		}
		ft.Target, err = declareAndDefineType(ast.Type, s, log)
		if err != nil {
			return nil, err
		}
		t := ft.Target
		targetIsPointer := false
		// The target for a destructor is always a pointer to the type it is destructing.
		if ast.TildeToken != nil {
			if _, ok := t.(*StructType); !ok {
				return nil, log.AddError(errlog.ErrorWrongTypeForDestructor, ast.Type.Location())
			}
			targetIsPointer = true
			ft.Target = &MutableType{Mutable: true, Type: &PointerType{Mode: PtrOwner, ElementType: ft.Target}}
		} else {
			if m, ok := t.(*MutableType); ok {
				t = m.Type
			}
			if ptr, ok := t.(*PointerType); ok {
				t = ptr.ElementType
				targetIsPointer = true
			}
		}
		if s.dualIsMut != -1 {
			// Do not register the dual function with its target type.
			switch target := t.(type) {
			case *StructType:
				if target.HasMember(f.name) {
					return nil, log.AddError(errlog.ErrorDuplicateScopeName, ast.Location(), f.name)
				}
				target.Funcs = append(target.Funcs, f)
			case *AliasType:
				if target.HasMember(f.name) {
					return nil, log.AddError(errlog.ErrorDuplicateScopeName, ast.Location(), f.name)
				}
				target.Funcs = append(target.Funcs, f)
			case *GenericInstanceType:
				if target.HasMember(f.name) {
					return nil, log.AddError(errlog.ErrorDuplicateScopeName, ast.Location(), f.name)
				}
				target.Funcs = append(target.Funcs, f)
			case *GenericType:
				if target.HasMember(f.name) {
					return nil, log.AddError(errlog.ErrorDuplicateScopeName, ast.Location(), f.name)
				}
				target.Funcs = append(target.Funcs, f)
				if err := parseFuncAttribs(ast, f, log); err != nil {
					return nil, err
				}
				// Do not inspect the function signature. This is done upon instantiation
				return f, nil
			default:
				return nil, log.AddError(errlog.ErrorTypeCannotHaveFunc, ast.Location())
			}
		}
		fixTargetStackOrder(ft, f.InnerScope, ast.Type.Location(), log)
		tthis := makeExprType(ft.Target)
		// If the target is not a pointer, then this is a value and it can be modified.
		// The same applies to destructors.
		if !targetIsPointer {
			tthis.Mutable = true
		}
		vthis := &Variable{name: "this", Type: tthis}
		f.InnerScope.AddElement(vthis, ast.Type.Location(), log)
	}
	f.Type.In, err = declareAndDefineParams(ast.Params, true, f.InnerScope, log)
	if err != nil {
		return nil, err
	}
	f.Type.Out, err = declareAndDefineParams(ast.ReturnParams, false, f.InnerScope, log)
	if err != nil {
		return nil, err
	}
	for i, p := range f.Type.In.Params {
		fixParameterStackOrder(ft, p, i, f.InnerScope, log)
	}
	for i, p := range f.Type.Out.Params {
		fixReturnStackOrder(ft, p, i, f.InnerScope, log)
	}
	if err := parseFuncAttribs(ast, f, log); err != nil {
		return nil, err
	}
	return f, nil
}

func fixParameterStackOrder(ft *FuncType, p *Parameter, pos int, s *Scope, log *errlog.ErrorLog) {
	if TypeHasPointers(p.Type) {
		// TODO: TF
		// et := NewExprType(p.Type)
		// p.Type = &GroupedType{GroupSpecifier: g, Type: p.Type, TypeBase: TypeBase{location: p.Location, component: p.Type.Component(), pkg: p.Type.Package()}}
	}
}

func fixReturnStackOrder(ft *FuncType, p *Parameter, pos int, s *Scope, log *errlog.ErrorLog) {
	if TypeHasPointers(p.Type) {
		// TODO: TF
		// et := NewExprType(p.Type)
		// p.Type = &GroupedType{GroupSpecifier: g, Type: p.Type, TypeBase: TypeBase{location: p.Location, component: p.Type.Component(), pkg: p.Type.Package()}}
	}
}

func fixTargetStackOrder(ft *FuncType, s *Scope, loc errlog.LocationRange, log *errlog.ErrorLog) {
	if TypeHasPointers(ft.Target) {
		// TODO: TF
		// et := NewExprType(ft.Target)
		// ft.Target = &GroupedType{GroupSpecifier: g, Type: ft.Target, TypeBase: TypeBase{location: loc, component: ft.Component(), pkg: ft.Package()}}
	}
}

func declareExternFunction(ast *parser.ExternFuncNode, s *Scope, log *errlog.ErrorLog) (*Func, error) {
	ft := &FuncType{TypeBase: TypeBase{name: ast.NameToken.StringValue, location: ast.Location(), pkg: s.PackageScope().Package}}
	f := &Func{name: ast.NameToken.StringValue, Type: ft, Ast: nil, OuterScope: s, Location: ast.Location(), IsExtern: true}
	if err := parseExternFuncAttribs(ast, f, log); err != nil {
		return nil, err
	}
	p, err := declareAndDefineParams(ast.Params, true, s, log)
	if err != nil {
		return nil, err
	}
	ft.In = p
	p, err = declareAndDefineParams(ast.ReturnParams, false, s, log)
	if err != nil {
		return nil, err
	}
	ft.Out = p
	return f, nil
}
