package bot

import (
	"bufio"
	"fmt"
	"os"

	gradix "github.com/armon/go-radix"
)

type Checker struct {
	tree *gradix.Tree
}

func (c *Checker) BadUserAgent(ua string) bool {
	_, ok := c.tree.Get(ua)
	return ok
}

func NewCheckerFromFile(file string) (Checker, error) {

	bots, err := os.Open(file)
	if err != nil {
		return Checker{}, fmt.Errorf("on open file: %v", err)
	}
	defer bots.Close()

	btree := gradix.New()

	scanner := bufio.NewScanner(bots)
	for scanner.Scan() {
		_, _ = btree.Insert(scanner.Text(), true)
	}
	if err := scanner.Err(); err != nil {
		return Checker{}, fmt.Errorf("on scan bots: %v", err)
	}
	return Checker{
		tree: btree,
	}, nil
}
