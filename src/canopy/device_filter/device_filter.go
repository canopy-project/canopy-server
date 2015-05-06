/*
 * Copyright 2015 Canopy Services, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package device_filter

import (
    "canopy/cloudvar"
    "canopy/datalayer"
    "fmt"
    "strings"
    "strconv"
)

type TokenTypeEnum int
const (
    TOKEN_COMPARE_OP TokenTypeEnum = iota
    TOKEN_BINARY_OP
    TOKEN_UNARY_OP
    TOKEN_FLOAT_VALUE
    TOKEN_STRING_VALUE
    TOKEN_BOOLEAN_VALUE
    TOKEN_OPEN_PAREN
    TOKEN_CLOSE_PAREN
)

type BinaryOpEnum int
const (
    AND BinaryOpEnum = iota
    OR
    LT
    LTE
    EQ
    NEQ
    GT
    GTE
)

type UnaryOpEnum int
const (
    NOT UnaryOpEnum = iota
    HAS
)

type Token struct {
    token_type TokenTypeEnum
    binary_op BinaryOpEnum
    unary_op UnaryOpEnum
    boolean_value bool
    float_value float64
    string_value string
}

type Expression interface {
    Value(device datalayer.Device) (cloudvar.CloudVarValue, error)
}

type DeviceFilter struct {
    expr Expression
}

// A BinaryOpExpression is a binary tree node that represents an binary
// operation.
//
//              AND
//             /   \
//     (operand0)  (operand1)
//
type BinaryOpExpression struct {
    operation BinaryOpEnum
    operand0 Expression
    operand1 Expression
}

// A UnaryOpExpression is a link that represents a unary operation:
//
//          NOT
//           |
//        (operand)
//
type UnaryOpExpression struct {
    operation UnaryOpEnum
    operand Expression
}

// A PropertyExpression is a leaf node that references a cloud variable or
// other device property value.
//
//          |
//      "temperature"
//
type PropertyExpression struct {
    property string
}

// An ImmediateExpression is a leaf node that contains a constant value
//
//          |
//         50.4
//
type ImmediateExpression struct {
    value cloudvar.CloudVarValue
}

func operandTokenToExpression(tok *Token) (Expression, error) {
    return nil, fmt.Errorf("Not implemented")
}


// BASIC TERMS:
// temperature > 50.4
// 4 = 2
// temperature != humidity
//
// BOOLEAN OPS:
//
// temprature > 50.4 AND 4 = 2 AND temperature != humidity

// multiSplit
//
// Takes a list of text strings <texts> and a list of separators <seps>, and
// returns a single flat list of strings which is similar to <texts> but has
// been further split up.
//
// Unlike strings.Split, the separators are not discarded.
//
// For example:
//
//  {
//      "Good news, everyone! ",
//      "There's a report on TV with some very bad news!"
//  }
//
//  Split on: {" ", "!"}
//
//  Becomes:
//  {
//      "Good",
//      " ",
//      "news",
//      " ",
//      "everyone",
//      "!",
//      " ",
//      "There's",
//      " ",
//      "a",
//      " ",
//      "report",
//      " ",
//      "on",
//      " ",
//      "TV",
//      " ",
//      "with",
//      " ",
//      "some",
//      " ",
//      "very",
//      " ",
//      "bad",
//      " ",
//      "news",
//      "!",
//  }
//
func multiSplit(texts []string, seps []string) []string {
    // This recursive implementation splits on seps[0] and then recursively
    // calls itself with a smaller seps list.

    if len(seps) == 0 {
        // base case
        return texts;
    }

    // recursive case, split on seps[0]
    out := []string{}
    for _, text := range texts {
        parts := strings.Split(text, seps[0])
        for i, part := range parts {
            if i > 0 {
                out = append(out, seps[0])
            }
            out = append(out, part)
        }
    }

    // recurse
    return multiSplit(out, seps[1:])
}

func isNumber(s string) bool {
    _, err := strconv.ParseFloat(s, 64)
    return err != nil
}

// Temperature > 64 AND humidity < 23
func stringToToken(s string) *Token {
    switch {
    case s == "=":
        return &Token{ token_type: TOKEN_BINARY_OP, binary_op: EQ }
    case s == "!=":
        return &Token{ token_type: TOKEN_BINARY_OP, binary_op: NEQ }
    case s == "<":
        return &Token{ token_type: TOKEN_BINARY_OP, binary_op: LT }
    case s == "<=":
        return &Token{ token_type: TOKEN_BINARY_OP, binary_op: LTE }
    case s == ">":
        return &Token{ token_type: TOKEN_BINARY_OP, binary_op: GT }
    case s == ">=":
        return &Token{ token_type: TOKEN_BINARY_OP, binary_op: GTE }
    case s == "(":
        return &Token{ token_type: TOKEN_OPEN_PAREN }
    case s == ")":
        return &Token{ token_type: TOKEN_CLOSE_PAREN }
    case s == "AND":
        return &Token{ token_type: TOKEN_BINARY_OP, binary_op: AND }
    case s == "OR":
        return &Token{ token_type: TOKEN_BINARY_OP, binary_op: OR }
    case s == "NOT":
        return &Token{ token_type: TOKEN_UNARY_OP, unary_op: NOT }
    case s == "HAS":
        return &Token{ token_type: TOKEN_UNARY_OP, unary_op: HAS }
    case strings.Trim(s, " ") == "":
        return nil
    case strings.ToLower(s) == "true":
        return &Token{ token_type: TOKEN_BOOLEAN_VALUE, boolean_value: true }
    case strings.ToLower(s) == "false":
        return &Token{ token_type: TOKEN_BOOLEAN_VALUE, boolean_value: false }
    case isNumber(s):
        fval, _ := strconv.ParseFloat(s, 64)
        return &Token{ token_type: TOKEN_FLOAT_VALUE, float_value: fval }
    default:
        return &Token{ token_type: TOKEN_STRING_VALUE, string_value: s }
    }
}

func tokenize(expr string) []*Token {
    out := []*Token{}
    tokStrings := multiSplit([]string{expr}, []string{" ", "(", ")"})
    for _, tokString := range tokStrings {
        tok := stringToToken(tokString)
        if tok != nil {
            out = append(out, tok)
        }
    }
    return out
}

func operatorHasPrecedence(token0, token1 *Token) bool {
    // for now no OoO
    return true
}

func infixToPrefix(tokens []*Token) ([]*Token, error) {
    stack := []*Token{}
    postfix := []*Token{}
    out := []*Token{}

    for _, token := range tokens {
        switch token.token_type {
        case TOKEN_BOOLEAN_VALUE, TOKEN_STRING_VALUE, TOKEN_FLOAT_VALUE:
            // If the scanned token is an operand, add it to the postfix array.
            postfix = append(postfix, token)

        case TOKEN_COMPARE_OP:
            // If the scanned token is an operator...

            if len(stack) == 0 {
                // ... and if the stack is empty, push the token to the stack.
                stack = append(stack, token)
            } else {
                // ... and if the stack is not empty, compare the precedence of
                // the operator with the operator on top of the stack.  If
                // topmost has higher precedence, pop the stack, else push the
                // scanned token to the stack.  Repeast as long as stack is not
                // empty and topmost has precedence of ther scanned token.
                for len(stack) > 0 && operatorHasPrecedence(token, stack[len(stack)-1]) {
                    // pop the stack and push the popped element to postfix
                    // array
                    popped := stack[len(stack)-1]
                    stack = stack[0:]
                    postfix = append(postfix, popped)
                }

                // push the scanned token to the stack
                stack = append(stack, token)
            }

        default:
            return nil, fmt.Errorf("Unexpected var")
        }
    }

    // Flush out anything remaining on stack
    for i := len(stack) - 1; i >= 0; i-- {
        popped := stack[i]
        postfix = append(postfix, popped)
    }

    // Reverse postfix to get prefix
    for i := len(postfix) - 1; i >= 0; i-- {
        out = append(out, postfix[i])
    }

    return out, nil
}

// Recursively convert prefix-ordered token array into filter tree structure.
func genExpressionTree(prefix []*Token) (Expression, []*Token, error) {
    
    // base case: token array is empty
    if len(prefix) == 0 {
        return nil, prefix, nil
    }

    token := prefix[0]
    switch token.token_type {
    case TOKEN_BOOLEAN_VALUE, TOKEN_STRING_VALUE, TOKEN_FLOAT_VALUE:
        // Token is an operand.
        outPrefix := prefix[1:]
        filter, err := operandTokenToExpression(token)
        return filter, outPrefix, err
    case TOKEN_BINARY_OP:
        // Recursive case.  Token is an binary operator.  Construct two child
        // filter trees as operands.
        operand0, newPrefix, err := genExpressionTree(prefix[1:])
        if err != nil {
            return nil, nil, err
        }
        operand1, newPrefix, err := genExpressionTree(newPrefix)
        if err != nil {
            return nil, nil, err
        }
        expr :=  &BinaryOpExpression{token.binary_op, operand0, operand1}
        return expr, newPrefix, err
    case TOKEN_UNARY_OP:
        // Recursive case.  Token is an unary operator.  Construct one child
        // filter tree as operand.
        operand, newPrefix, err := genExpressionTree(prefix[1:])
        if err != nil {
            return nil, nil, err
        }
        expr := &UnaryOpExpression{token.unary_op, operand}
        return expr, newPrefix, err
    default:
        return nil, nil, fmt.Errorf("Unexpected token")
    }
}

func Parse(expr string) (Expression, error) {
    // TOKENIZE
    tokens := tokenize(expr)

    // CONVERT TO PREFIX
    prefixTokens, err := infixToPrefix(tokens)

    // PARSE INTO TO FILTER TREE
    expression, _, err := genExpressionTree(prefixTokens)
    return expression, err
}


func (expr *BinaryOpExpression)Value(device datalayer.Device) (cloudvar.CloudVarValue, error) {
    switch expr.operation {
    case AND, OR:
        v0, err := expr.operand0.Value(device)
        if err != nil {
            return nil, err
        }
        v0Bool, ok := v0.(bool)
        if !ok {
            return nil, fmt.Errorf("Left operand must evaluate to boolean")
        }

        v1, err := expr.operand1.Value(device)
        if err != nil {
            return nil, err
        }
        v1Bool, ok := v1.(bool)
        if !ok {
            return nil, fmt.Errorf("Right operand must evaluate to boolean")
        }

        if (expr.operation == AND) {
            return (v0Bool && v1Bool), nil
        } else {
            return (v0Bool || v1Bool), nil
        }

    case LT, LTE, EQ, NEQ, GT, GTE:
        mapping := map[BinaryOpEnum]cloudvar.CompareOpEnum{
            LT: cloudvar.LT,
            LTE: cloudvar.LTE,
            EQ: cloudvar.EQ,
            NEQ: cloudvar.NEQ,
            GT: cloudvar.GT,
            GTE: cloudvar.GTE,
        }
        v0, err := expr.operand0.Value(device)
        if err != nil {
            return false, err
        }
        v1, err := expr.operand1.Value(device)
        if err != nil {
            return false, err
        }
        return cloudvar.CompareValues(v0, v1, mapping[expr.operation])
    default:
        return false, fmt.Errorf("Unexpected binary operation")
    }
}

func (expr *ImmediateExpression)Value(device datalayer.Device) (cloudvar.CloudVarValue, error) {
    return expr.value, nil
}

func (expr *PropertyExpression)Value(device datalayer.Device) (cloudvar.CloudVarValue, error) {
    sample, err := device.LatestDataByName(expr.property)
    if err != nil {
        return nil, err
    }
    return sample.Value, nil
}

func (expr *UnaryOpExpression)Value(device datalayer.Device) (cloudvar.CloudVarValue, error) {
    switch expr.operation {
    case NOT:
        val, err := expr.operand.Value(device)
        if err != nil {
            return nil, err
        }
        valBool, ok := val.(bool)
        if !ok{
            return nil, fmt.Errorf("NOT operand must evaluate to boolean")
        }
        return !valBool, nil
    case HAS:
        propExpr, ok := expr.operand.(*PropertyExpression)
        if !ok {
            return nil, fmt.Errorf("HAS operand must be a variable reference")
        }
        _, err := device.LookupVarDef(propExpr.property)
        return (err == nil), nil
    default:
        return false, fmt.Errorf("Unexpected unary operation")
    }
}
