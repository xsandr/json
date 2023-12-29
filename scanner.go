package json

const (
	ObjectStart = '{' // {
	ObjectEnd   = '}' // }
	String      = '"' // "
	Colon       = ':' // :
	Comma       = ',' // ,
	ArrayStart  = '[' // [
	ArrayEnd    = ']' // ]
	True        = 't' // t
	False       = 'f' // f
	Null        = 'n' // n
)

// NewScanner returns a new Scanner for given []byte
// A Scanner produces a stream of tokens
func NewScanner(data []byte) *Scanner {
	return &Scanner{
		data: data,
	}
}

// Scanner implements a JSON scanner as defined in RFC 7159.
type Scanner struct {
	data   []byte
	offset int
}

var whitespace = [256]bool{
	' ':  true,
	'\r': true,
	'\n': true,
	'\t': true,
}

var openArray = [256]bool{
	'[': true,
}
var closeArray = [256]bool{
	']': true,
}

var openObject = [256]bool{
	'{': true,
}
var closeObject = [256]bool{
	'}': true,
}

// Next returns a []byte referencing the next lexical token in the stream.
// The []byte is valid until Next is called again.
// If the stream is at its end, or an error has occurred, Next returns a zero
// length []byte slice.
//
// A valid token begins with one of the following:
//
//	{ Object start
//	[ Array start
//	} Object end
//	] Array End
//	, Literal comma
//	: Literal colon
//	t JSON true
//	f JSON false
//	n JSON null
//	" A string, possibly containing backslash escaped entites.
//	-, 0-9 A number
func (s *Scanner) Next() []byte {
	if s.offset > len(s.data)-1 {
		return nil
	}
	w := s.data[s.offset:]
	initialOffset := s.offset
	for {
		for pos, c := range w {
			// strip any leading whitespace.
			if whitespace[c] {
				continue
			}

			// simple case
			switch c {
			case ObjectStart, ObjectEnd, Colon, Comma, ArrayStart, ArrayEnd:
				s.offset += pos + 1
				return w[pos : pos+1]
			}
			s.offset = initialOffset + pos

			switch c {
			case True:
				s.offset += s.validateToken("true")
			case False:
				s.offset += s.validateToken("false")
			case Null:
				s.offset += s.validateToken("null")
			case String:
				length := s.parseString()
				if length < 2 {
					return nil
				}
				s.offset += length

			default:
				// ensure the number is correct.
				s.offset += s.parseNumber(c)
			}
			return s.data[initialOffset+pos : s.offset]
		}

		s.offset += len(w)
		w = s.data[s.offset:]
		if len(w) == 0 {
			// eof
			return nil
		}
	}
}

func (s *Scanner) skipArray() {
	w := s.data[s.offset:]
	count := 1
	inString := false
	escaped := false

	for i, c := range w {
		if c == '"' && !inString {
			inString = true
			continue
		}

		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '"':
				inString = false
			case c == '\\':
				escaped = true
			}
			continue
		}

		if openArray[c] {
			count++
		}
		if closeArray[c] {
			count--
			if count == 0 {
				s.offset += i + 1
				return
			}
		}
	}

	s.offset += len(w) + 1
}

func (s *Scanner) skipObject() {
	w := s.data[s.offset:]
	count := 1
	inString := false
	escaped := false

	for i, c := range w {
		if c == '"' && !inString {
			inString = true
			continue
		}

		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '"':
				inString = false
			case c == '\\':
				escaped = true
			}
			continue
		}

		if openObject[c] {
			count++
		}
		if closeObject[c] {
			count--
			if count == 0 {
				s.offset += i + 1
				return
			}
		}
	}
	s.offset += len(w) + 1
}

func (s *Scanner) validateToken(expected string) int {
	w := s.data[s.offset:]
	n := len(expected)
	if len(w) >= n {
		if string(w[:n]) != expected {
			// doesn't match
			return 0
		}
		return n
	}
	return 0
}

// parseString returns the length of the string token
// located at the start of the window or 0 if there is no closing " before the end of the data
func (s *Scanner) parseString() int {
	escaped := false
	w := s.data[s.offset+1:]
	offset := 0
	for _, c := range w {
		offset++
		switch {
		case escaped:
			escaped = false
		case c == '"':
			// finished
			return offset + 1
		case c == '\\':
			escaped = true
		}
	}
	// no closing "
	return 0
}

func (s *Scanner) parseNumber(c byte) int {
	const (
		begin = iota
		leadingzero
		anydigit1
		decimal
		anydigit2
		exponent
		expsign
		anydigit3
	)

	offset := 0
	w := s.data[s.offset:]
	// w := s.data[s.offset:]
	// int vs uint8 costs 10% on canada.json
	var state uint8 = begin

	// handle the case that the first character is a hyphen
	if c == '-' {
		offset++
	}

	for {
		for _, elem := range w[offset:] {
			switch state {
			case begin:
				if elem >= '1' && elem <= '9' {
					state = anydigit1
				} else if elem == '0' {
					state = leadingzero
				} else {
					// error
					return 0
				}
			case anydigit1:
				if elem >= '0' && elem <= '9' {
					// stay in this state
					break
				}
				fallthrough
			case leadingzero:
				if elem == '.' {
					state = decimal
					break
				}
				if elem == 'e' || elem == 'E' {
					state = exponent
					break
				}
				return offset // finished.
			case decimal:
				if elem >= '0' && elem <= '9' {
					state = anydigit2
				} else {
					// error
					return 0
				}
			case anydigit2:
				if elem >= '0' && elem <= '9' {
					break
				}
				if elem == 'e' || elem == 'E' {
					state = exponent
					break
				}
				return offset // finished.
			case exponent:
				if elem == '+' || elem == '-' {
					state = expsign
					break
				}
				fallthrough
			case expsign:
				if elem >= '0' && elem <= '9' {
					state = anydigit3
					break
				}
				// error
				return 0
			case anydigit3:
				if elem < '0' || elem > '9' {
					return offset
				}
			}
			offset++
		}

		w = s.data[offset:]
		if len(w) == 0 {
			// end of the item. However, not necessarily an error. Make
			// sure we are in a state that allows ending the number.
			switch state {
			case leadingzero, anydigit1, anydigit2, anydigit3:
				return offset
			default:
				// error otherwise, the number isn't complete.
				return 0
			}
		}
	}
	return offset
}
