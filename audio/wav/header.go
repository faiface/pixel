package wav

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

type header struct {
	fileSize      int32
	formatSize    int32
	formatType    int16
	numChans      int16
	sampleRate    int32
	byteRate      int32
	bytesPerFrame int16
	bitsPerSample int16
	dataSize      int32
}

func readHeader(r io.Reader) (header, error) {
	var (
		h  header
		er errReader
	)
	err := er.
		ReadString(r, "RIFF", errors.New("missing RIFF at the beginning")).
		ReadBinary(r, binary.LittleEndian, &h.fileSize).
		ReadString(r, "WAVE", errors.New("unsupported file type")).
		ReadString(r, "fmt\x00", errors.New("missing format chunk marker")).
		ReadBinary(r, binary.LittleEndian, &h.formatSize).
		ReadBinary(r, binary.LittleEndian, &h.formatType).
		ReadBinary(r, binary.LittleEndian, &h.numChans).
		ReadBinary(r, binary.LittleEndian, &h.sampleRate).
		ReadBinary(r, binary.LittleEndian, &h.byteRate).
		ReadBinary(r, binary.LittleEndian, &h.bytesPerFrame).
		ReadBinary(r, binary.LittleEndian, &h.bitsPerSample).
		ReadString(r, "data", errors.New("missing data chunk marker")).
		ReadBinary(r, binary.LittleEndian, &h.dataSize).
		Err()
	return h, err
}

type errReader struct {
	err error
}

func (e *errReader) ReadString(r io.Reader, s string, notThereErr error) *errReader {
	if e.err != nil {
		return e
	}
	buf := make([]byte, len(s))
	_, err := r.Read(buf)
	if err != nil {
		e.err = errors.Wrap(err, "error while reading header")
		return e
	}
	if string(buf) != s {
		e.err = errors.Wrap(err, "invalid header")
	}
	return e
}

func (e *errReader) ReadBinary(r io.Reader, order binary.ByteOrder, data interface{}) *errReader {
	if e.err != nil {
		return e
	}
	err := binary.Read(r, order, data)
	if err != nil {
		e.err = errors.Wrap(err, "invalid header")
	}
	return e
}

func (e *errReader) Err() error {
	return e.err
}
