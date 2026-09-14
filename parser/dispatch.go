package parser

import (
	"cool/ast"
	"cool/token"
)

func (p *Parser) parseArgs() ([]ast.Expr, error) {
	if _, err := p.expect(token.LPAREN); err != nil {
		return nil, err
	}
	var args []ast.Expr
	if !p.check(token.RPAREN) {
		for {
			e, err := p.parseExpr()
			if err != nil {
				return nil, err
			}
			args = append(args, e)
			if p.check(token.COMMA) {
				p.next()
				continue
			}
			break
		}
	}
	if _, err := p.expect(token.RPAREN); err != nil {
		return nil, err
	}
	return args, nil
}

func (p *Parser) parseDispatch() (ast.Expr, error) {
	var recv ast.Expr

	if p.check(token.OBJECTID) {
		idTok := p.next()
		if p.check(token.LPAREN) {
			args, err := p.parseArgs()
			if err != nil {
				return nil, err
			}
			return &ast.Dispatch{Receiver: nil, Method: idTok.Literal, Args: args, Line: idTok.Line, Column: idTok.Column}, nil
		}
		recv = &ast.Object{Name: idTok.Literal, Line: idTok.Line, Column: idTok.Column}
	} else {
		r, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		recv = r
	}

	for {
		if p.check(token.DOT) {
			dotTok := p.next()
			_ = dotTok
			mTok, err := p.expect(token.OBJECTID)
			if err != nil {
				return nil, err
			}
			args, err := p.parseArgs()
			if err != nil {
				return nil, err
			}
			rl, rc := recvPos(recv)
			recv = &ast.Dispatch{Receiver: recv, Method: mTok.Literal, Args: args, Line: rl, Column: rc}
			continue
		}
		if p.check(token.AT) {
			p.next()
			tTok, err := p.expect(token.TYPEID)
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(token.DOT); err != nil {
				return nil, err
			}
			mTok, err := p.expect(token.OBJECTID)
			if err != nil {
				return nil, err
			}
			args, err := p.parseArgs()
			if err != nil {
				return nil, err
			}
			rl, rc := recvPos(recv)
			recv = &ast.StaticDispatch{Receiver: recv, StaticType: tTok.Literal, Method: mTok.Literal, Args: args, Line: rl, Column: rc}
			continue
		}
		break
	}
	return recv, nil
}

func recvPos(e ast.Expr) (int, int) {
	switch n := e.(type) {
	case *ast.Object:
		return n.Line, n.Column
	case *ast.IntConst:
		return n.Line, n.Column
	case *ast.StringConst:
		return n.Line, n.Column
	case *ast.BoolConst:
		return n.Line, n.Column
	case *ast.New:
		return n.Line, n.Column
	case *ast.IsVoid:
		return n.Line, n.Column
	case *ast.Dispatch:
		return n.Line, n.Column
	case *ast.StaticDispatch:
		return n.Line, n.Column
	case *ast.Case:
		return n.Line, n.Column
	default:
		return 0, 0
	}
}
