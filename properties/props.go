package properties

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Props map[string]string

func Load(path string) (Props, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	p := Props{}
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '='); i >= 0 {
			k := strings.TrimSpace(line[:i])
			v := strings.TrimSpace(line[i+1:])
			p[k] = v
		}
	}
	return p, s.Err()
}

func (p Props) Int(key string, def int) int {
	if v, ok := p[key]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
