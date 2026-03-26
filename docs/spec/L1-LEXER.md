# LAYER 1: LEXER
**Status:** FROZEN
**Depends on:** L0-AXIOMS.md

---

## PURPOSE

The lexer converts source text (UTF-8) into a sequence of tokens.
Every other layer operates on tokens, never on raw characters.
These rules are FROZEN: a program's token sequence never changes.

---

### RULE LEX-001: Source Encoding

**Status:** FROZEN
**Layer:** LEX
**Depends on:** none

Candor source files are UTF-8 encoded text. A byte-order mark (BOM)
at the start of a file is ignored. Source files must not contain null bytes.

**Invariant:** All source characters are valid UTF-8 code points.

**Compliance Test:**
PASS: Any valid UTF-8 file
FAIL: A file with a raw ISO-8859-1 byte above 0x7F that is not valid UTF-8

---

### RULE LEX-002: Tokenization Is Maximal Munch

**Status:** FROZEN
**Layer:** LEX
**Depends on:** LEX-001

At each position, the lexer consumes the longest possible token.
Ambiguity between two valid tokenizations is resolved by choosing the longer token.

**Compliance Test:**
PASS: `!=` is tokenized as one `TokBangEq` token, not `TokBang` + `TokEq`
PASS: `<<` is tokenized as `TokLShift`, not two `TokLt` tokens

---

### RULE LEX-005: Whitespace

**Status:** FROZEN
**Layer:** LEX
**Depends on:** LEX-001

Whitespace characters (space 0x20, tab 0x09, carriage return 0x0D,
newline 0x0A) are token separators. They produce no tokens.
Leading and trailing whitespace on any line is ignored.

**Invariant:** No token contains whitespace, except inside string literals (LEX-023).

**Compliance Test:**
PASS: `let   x  =  5` tokenizes identically to `let x = 5`

---

### RULE LEX-010: Token Kinds

**Status:** FROZEN
**Layer:** LEX
**Depends on:** LEX-002

The complete set of token kinds is:

**Literals:**
| Kind | Description | Example |
|------|-------------|---------|
| `TokInt` | Integer literal | `42`, `0xFF`, `0b1010`, `0o17` |
| `TokFloat` | Float literal | `3.14`, `1.5e-3` |
| `TokStr` | String literal (with quotes) | `"hello"` |
| `TokTrue` | Boolean true | `true` |
| `TokFalse` | Boolean false | `false` |
| `TokUnit` | Unit literal | `unit` |

**Identifiers:**
| Kind | Description |
|------|-------------|
| `TokIdent` | User-defined name |

**Operators:**
| Kind | Lexeme | Kind | Lexeme |
|------|--------|------|--------|
| `TokPlus` | `+` | `TokMinus` | `-` |
| `TokStar` | `*` | `TokSlash` | `/` |
| `TokPercent` | `%` | `TokAmp` | `&` |
| `TokPipe` | `\|` | `TokCaret` | `^` |
| `TokTilde` | `~` | `TokBang` | `!` |
| `TokLt` | `<` | `TokGt` | `>` |
| `TokLtEq` | `<=` | `TokGtEq` | `>=` |
| `TokEqEq` | `==` | `TokBangEq` | `!=` |
| `TokAmpAmp` | `&&` | `TokPipePipe` | `\|\|` |
| `TokLShift` | `<<` | `TokRShift` | `>>` |
| `TokPlusEq` | `+=` | `TokMinusEq` | `-=` |
| `TokStarEq` | `*=` | `TokSlashEq` | `/=` |
| `TokPercentEq` | `%=` | `TokEq` | `=` |
| `TokArrow` | `->` | `TokFatArrow` | `=>` |
| `TokColonColon` | `::` | `TokDotDot` | `..` |

**Delimiters:**
| Kind | Lexeme | Kind | Lexeme |
|------|--------|------|--------|
| `TokLParen` | `(` | `TokRParen` | `)` |
| `TokLBrace` | `{` | `TokRBrace` | `}` |
| `TokLBracket` | `[` | `TokRBracket` | `]` |
| `TokColon` | `:` | `TokSemicolon` | `;` |
| `TokComma` | `,` | `TokDot` | `.` |

**Special:**
| Kind | Description |
|------|-------------|
| `TokEOF` | End of input |
| `TokDirective` | `#word` (e.g., `#intent`, `#c_header`) |

**Keywords:** see LEX-050.

---

### RULE LEX-021: Integer Literals

**Status:** FROZEN
**Layer:** LEX
**Depends on:** LEX-002

Integer literals match one of these forms:

```ebnf
IntLit   ::= DecInt | HexInt | OctInt | BinInt
DecInt   ::= [1-9] [0-9_]* | '0'
HexInt   ::= '0x' [0-9A-Fa-f_]+
OctInt   ::= '0o' [0-7_]+
BinInt   ::= '0b' [01_]+
```

Underscore `_` may appear as a visual separator between digits but
not as the first or last character of the numeric portion.

The lexeme of an integer literal is the full matched text including prefix.
The numeric value is the integer the literal represents.
Underscores do not contribute to the value.

**Invariant:** Integer literals do not carry a type. Type assignment
is done at Layer 3 (TYP-091).

**Compliance Test:**
PASS: `0xFF` → value 255, kind `TokInt`
PASS: `1_000_000` → value 1000000
FAIL: `_42` → not an integer literal (leading underscore makes it an identifier)
FAIL: `42_` → trailing underscore is invalid

---

### RULE LEX-022: Float Literals

**Status:** FROZEN
**Layer:** LEX
**Depends on:** LEX-021

```ebnf
FloatLit ::= DecInt '.' DecInt? ExponentPart?
           | DecInt ExponentPart
ExponentPart ::= ('e' | 'E') ('+' | '-')? DecInt
```

A float literal must contain either a `.` or an exponent part (or both).
`42` is an integer literal. `42.0` is a float literal. `42e0` is a float literal.

**Invariant:** Float literals do not carry a type. Type assignment is done at TYP-092.

**Compliance Test:**
PASS: `3.14` → float literal
PASS: `1e10` → float literal
FAIL: `3.` → invalid (requires digit after decimal point)

---

### RULE LEX-023: String Literals

**Status:** FROZEN
**Layer:** LEX
**Depends on:** LEX-001, LEX-002

```ebnf
StrLit ::= '"' StrChar* '"'
StrChar ::= [^"\\] | EscapeSeq
EscapeSeq ::= '\\' ( 'n' | 't' | 'r' | '"' | '\\' | '0' | UnicodeEsc )
UnicodeEsc ::= 'u{' HexDigit+ '}'
```

The **lexeme** of a string literal includes the surrounding double-quote characters.
The **value** of a string literal is the sequence of Unicode code points
after escape sequence expansion.

Multi-line strings are not supported at this layer. A string literal
must start and end on the same line.

**Compliance Test:**
PASS: `"hello"` → lexeme `"hello"`, value `hello`
PASS: `"line\n"` → value is `line` followed by newline
FAIL: A string spanning two lines without escape

---

### RULE LEX-024: String Escape Sequences

**Status:** FROZEN
**Layer:** LEX
**Depends on:** LEX-023

| Escape | Unicode Code Point | Character |
|--------|-------------------|-----------|
| `\\n` | U+000A | Line Feed |
| `\\t` | U+0009 | Horizontal Tab |
| `\\r` | U+000D | Carriage Return |
| `\\"` | U+0022 | Double Quote |
| `\\\\` | U+005C | Backslash |
| `\\0` | U+0000 | Null |
| `\\u{HHHH}` | U+HHHH | Unicode scalar value |

An unrecognized escape sequence is a lexer error.

**Compliance Test:**
PASS: `"tab:\there"` → value `tab:` + U+0009 + `here`
FAIL: `"\q"` → unrecognized escape sequence

---

### RULE LEX-030: Boolean Literals

**Status:** FROZEN
**Layer:** LEX
**Depends on:** LEX-050 (keywords)

The tokens `true` and `false` are keyword tokens of kind `TokTrue` and `TokFalse`.
They are not identifiers.

**Compliance Test:**
PASS: `true` → kind `TokTrue`
PASS: `false` → kind `TokFalse`
FAIL: `true` bound as an identifier name via `let true = 5` → parse error

---

### RULE LEX-040: Identifiers

**Status:** FROZEN
**Layer:** LEX
**Depends on:** LEX-002, LEX-050

```ebnf
Identifier ::= (Letter | '_') (Letter | Digit | '_')*
Letter     ::= [a-zA-Z]
Digit      ::= [0-9]
```

An identifier may not be a keyword (see LEX-050).
Identifiers beginning with `_` are valid; a single `_` is the wildcard pattern.
Identifiers are case-sensitive: `Foo` and `foo` are different.

**Invariant:** No identifier is a keyword.

**Compliance Test:**
PASS: `foo`, `_bar`, `x42`, `MyStruct`
FAIL: `let` → keyword, not an identifier
FAIL: `42foo` → starts with digit

---

### RULE LEX-050: Keywords

**Status:** FROZEN
**Layer:** LEX
**Depends on:** LEX-040

The following are reserved keywords and may not be used as identifiers:

```
as       assert   box      break    cap      const    continue
effects  else     enum     ensures  extern   f16      f32
f64      false    fn       for      forall   if       impl
in       let      loop     match    module   mut      none
not      ok       option   pure     ref      refmut   requires
result   return   ring     secret   self     some     spawn
str      struct   task     tensor   trait    true     unit
use      vec      while
```

**Compliance Test:**
PASS: Tokenizing `fn` → `TokFn` keyword token
FAIL: Using `fn` as a variable name `let fn = 5`

---

### RULE LEX-080: Comments

**Status:** FROZEN
**Layer:** LEX
**Depends on:** LEX-001

Two comment forms are recognized:

**Line comments:** `##` followed by any characters to end of line.
Doc comments: `///` followed by any characters to end of line.
Line comments and doc comments produce no tokens.
Doc comment content is available to the doc generator (Layer DOC).

```ebnf
LineComment ::= '##' [^\n]* '\n'
DocComment  ::= '///' [^\n]* '\n'
```

**Note:** `//` is NOT a comment in Candor. Use `##`.

**Compliance Test:**
PASS: `let x = 5 ## this is a comment` → tokenizes as `let x = 5`
PASS: `/// doc comment` → no token produced
FAIL: `// not a comment` → tokenized as `TokSlash TokSlash ...`

---

### RULE LEX-090: Directives

**Status:** FROZEN
**Layer:** LEX
**Depends on:** LEX-040

A directive is `#` immediately followed by an identifier (no space).

```ebnf
Directive ::= '#' Identifier
```

Directives produce a `TokDirective` token with the identifier portion as the lexeme
(not including the `#`).

Known directives: `#intent`, `#mcp_tool`, `#c_header`, `#test`, `#export_json`.
Unknown directives produce a warning, not an error.

**Compliance Test:**
PASS: `#intent` → `TokDirective` with lexeme `intent`
PASS: `#unknown_future_directive` → `TokDirective` with warning
FAIL: `# intent` (space between `#` and name) → not a directive

---

*End of L1-LEXER.md*
