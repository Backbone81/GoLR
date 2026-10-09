package base

import (
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
)

const (
	HeaderContentLength = "Content-Length"
	HeaderContentType   = "Content-Type"
)

type ProtocolParser struct {
	reader io.Reader
	buffer []byte
	header map[string]string
	err    error
}

func NewProtocolParser(reader io.Reader) *ProtocolParser {
	return &ProtocolParser{
		reader: reader,
		buffer: make([]byte, 1024),
		header: make(map[string]string),
	}
}

//nolint:gocognit,cyclop,funlen
func (p *ProtocolParser) Next() bool {
	if p.err != nil {
		// When we encountered an error before, we do not continue parsing.
		return false
	}

	clear(p.header)
	var headerName, headerValue string
	state := 0
	for {
		p.buffer = p.buffer[:1]
		if _, p.err = io.ReadFull(p.reader, p.buffer); p.err != nil {
			if errors.Is(p.err, io.EOF) {
				if state == 0 && len(headerName) == 0 && len(p.header) == 0 {
					// An end of file before the first byte of a message is the only correct end of the stream.
					p.err = nil
				} else {
					p.err = io.ErrUnexpectedEOF
				}
			}
			return false
		}

		nextByte := p.buffer[0]
		if nextByte >= 0x80 {
			p.err = fmt.Errorf("non-ASCII byte 0x%02x in header", nextByte)
			return false
		}
		switch state {
		case 0:
			// Header name
			switch nextByte {
			case ':':
				if len(headerName) == 0 {
					p.err = errors.New("header field has an empty name")
					return false
				}
				state = 1
			case '\r':
				if len(headerName) > 0 {
					p.err = fmt.Errorf("header field %q has no ': ' separator", headerName)
					return false
				}
				state = 4
			default:
				headerName += string(nextByte)
			}
		case 1:
			// Header name/value separator
			if nextByte != ' ' {
				p.err = fmt.Errorf("header field %q: expected ' ' after ':', got %q", headerName, nextByte)
				return false
			}
			state = 2
		case 2:
			// Header value
			switch nextByte {
			case '\r':
				state = 3
			default:
				headerValue += string(nextByte)
			}
		case 3:
			// Header field terminator
			if nextByte != '\n' {
				p.err = fmt.Errorf("header field %q: expected '\\n' after '\\r', got %q", headerName, nextByte)
				return false
			}
			p.header[textproto.CanonicalMIMEHeaderKey(headerName)] = headerValue
			headerName = ""
			headerValue = ""
			state = 0
		case 4:
			// Header/content separator
			switch nextByte {
			case '\n':
				// We reached the end of the header. We need to read the content next.
				contentLength, ok := p.header[HeaderContentLength]
				if !ok {
					p.err = fmt.Errorf("required header %q is missing", HeaderContentLength)
					return false
				}

				typedContentLength, err := strconv.Atoi(contentLength)
				if err != nil {
					p.err = fmt.Errorf(
						"required header %q did not provide an integer, but %q: %w",
						HeaderContentLength,
						contentLength,
						err,
					)
					return false
				}
				if typedContentLength < 0 || 10*1024*1024 < typedContentLength {
					p.err = fmt.Errorf(
						"the value for required header %q was out of range (0-10 MB)",
						HeaderContentLength,
					)
					return false
				}

				if cap(p.buffer) < typedContentLength {
					p.buffer = make([]byte, typedContentLength)
				}
				p.buffer = p.buffer[:typedContentLength]

				if _, err := io.ReadFull(p.reader, p.buffer); err != nil {
					if errors.Is(err, io.EOF) {
						// The header announced content, so the end of the stream is never correct here.
						err = io.ErrUnexpectedEOF
					}
					p.err = fmt.Errorf("reading content: %w", err)
					return false
				}
				return true
			default:
				p.err = fmt.Errorf(
					"unexpected character %q in header/content separator, expected \\n",
					string(nextByte),
				)
				return false
			}
		}
	}
}

func (p *ProtocolParser) Err() error {
	return p.err
}

func (p *ProtocolParser) Header() map[string]string {
	return p.header
}

func (p *ProtocolParser) Content() []byte {
	return p.buffer
}
