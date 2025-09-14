package generator

import "github.com/dspasibenko/pargus/pkg/parser"

func bitMask(start, end int) uint64 {
	width := end - start + 1
	return (uint64(1)<<width - 1) << start
}

func flattenComments(cg *parser.CommentGroup) []string {
	if cg == nil {
		return nil
	}
	var out []string
	for _, e := range cg.Elements {
		if e.Comment != nil {
			out = append(out, *e.Comment)
		}
		if e.EmptyLine != nil {
			out = append(out, "")
		}
	}
	return out
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
