package comment

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Type int

const (
	Single Type = iota
	Block
)

func (t Type) String() string {
	switch t {
	case Single:
		return "single"
	case Block:
		return "block"
	default:
		panic("unknown")
	}
}

type Comment struct {
	Type       Type
	Start, End int
	Lines      []string
}

func (c *Comment) String() string {
	var r string
	if c.Start == c.End {
		r = strconv.Itoa(c.Start)
	} else {
		r = fmt.Sprintf("%d-%d", c.Start, c.End)
	}

	return fmt.Sprintf("%s:%s:%v", c.Type, r, c.Lines)
}

type Registry interface {
	GetSingle(line int) (string, bool)
	GetBlock(line int) *Comment
}

type NullRegistry struct{}

func (r *NullRegistry) GetSingle(line int) (string, bool) { return "", false }
func (r *NullRegistry) GetBlock(line int) *Comment        { return nil }

type registryImpl map[int][]*Comment

func (r registryImpl) GetSingle(line int) (string, bool) {
	arr, found := r[line]
	if !found {
		return "", false
	}

	for _, c := range arr {
		if c.Type == Single {
			return c.Lines[0], true
		}
	}

	return "", false
}

func (r registryImpl) GetBlock(line int) *Comment {
	arr, found := r[line]
	if !found {
		return nil
	}

	for _, c := range arr {
		if c.Type == Block {
			return c
		}
	}

	return nil
}

var (
	numberLinePattern  = regexp.MustCompile(`^(\d+)(?:-(\d+))?(?:\s+//\s+(\S.*))?$`)
	commentLinePattern = regexp.MustCompile(`^//\s+(\S.*)$`)
)

func matchCommentLine(line string) (string, bool) {
	parts := commentLinePattern.FindStringSubmatch(line)
	if parts == nil {
		return "", false
	}

	return parts[1], true
}

func Read(r io.Reader) (Registry, error) {
	out := registryImpl{}

	var curBlock *Comment

	scanner := bufio.NewScanner(r)
	for lineNum := 1; scanner.Scan(); lineNum++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if comment, found := matchCommentLine(line); found {
			if len(curBlock.Lines) == 0 {
				if _, found := out[curBlock.Start]; !found {
					out[curBlock.Start] = []*Comment{}
				}
				out[curBlock.Start] = append(out[curBlock.Start], curBlock)
			}
			curBlock.Lines = append(curBlock.Lines, comment)

		} else if parts := numberLinePattern.FindStringSubmatch(line); parts != nil {
			startStr, endStr, comment := parts[1], parts[2], parts[3]

			start, err := strconv.ParseUint(startStr, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("%d: bad start: %v", lineNum, err)
			}

			end := start
			if endStr != "" {
				var err error
				end, err = strconv.ParseUint(endStr, 10, 32)
				if err != nil {
					return nil, fmt.Errorf("%d: bad end: %v", lineNum, err)
				}
			}

			if _, found := out[int(start)]; found {
				return nil, fmt.Errorf("%d: duplicate start %v", lineNum, start)
			}

			if comment != "" {
				out[int(start)] = []*Comment{&Comment{Single, int(start), int(start), []string{comment}}}
			}

			curBlock = &Comment{Block, int(start), int(end), []string{}}

		} else {
			return nil, fmt.Errorf("%d: unable to parse", lineNum)
		}

	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func ReadFromPath(path string) (Registry, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	return Read(fp)
}
