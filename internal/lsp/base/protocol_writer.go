package base

import (
	"io"
	"strconv"
)

// WriteMessage writes the header and the content of a message.
func WriteMessage(writer io.Writer, content []byte) error {
	header := HeaderContentLength + ": " + strconv.Itoa(len(content)) + "\r\n\r\n"
	if _, err := io.WriteString(writer, header); err != nil {
		return err
	}
	_, err := writer.Write(content)
	return err
}
