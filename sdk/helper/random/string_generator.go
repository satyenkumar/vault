package random

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math"
	"time"
)

type Rule interface {
	Pass(value []rune) bool
}

type StringGenerator struct {
	Length  int    `mapstructure:"length"`
	Charset []rune `mapstructure:"charset"`
	Rules   []Rule `mapstructure:"-"`

	rng io.Reader
}

func (g *StringGenerator) Generate(ctx context.Context) (str string, err error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second) // Ensure there's a timeout on the context
	defer cancel()

LOOP:
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("timed out generating string")
		default:
			str, err = g.generate()
			if err != nil {
				return "", err
			}
			if str == "" {
				continue LOOP
			}
			return str, err
		}
	}
}

func (g *StringGenerator) generate() (str string, err error) {
	// If performance improvements need to be made, this can be changed to read a batch of
	// potential strings at once rather than one at a time. This will significantly
	// improve performance, but at the cost of added complexity.
	runes, err := randomRunes(g.rng, g.Charset, g.Length)
	if err != nil {
		return "", fmt.Errorf("unable to generate random characters: %w", err)
	}

	for _, rule := range g.Rules {
		if !rule.Pass(runes) {
			return "", nil
		}
	}

	// Passed all rules
	return string(runes), nil
}

// randomRunes creates a random string based on the provided charset. The charset is limited to 255 characters, but
// could be expanded if needed. Expanding the maximum charset size will decrease performance because it will need to
// combine bytes into a larger integer using binary.BigEndian.Uint16() function.
func randomRunes(rng io.Reader, charset []rune, length int) (randStr []rune, err error) {
	if len(charset) == 0 {
		return nil, fmt.Errorf("no charset specified")
	}
	if len(charset) > math.MaxUint8 {
		return nil, fmt.Errorf("charset is too long: limited to %d characters", math.MaxUint8)
	}
	if length <= 0 {
		return nil, fmt.Errorf("unable to generate a zero or negative length runeset")
	}

	if rng == nil {
		rng = rand.Reader
	}

	charsetLen := byte(len(charset))
	data := make([]byte, length)
	_, err = rng.Read(data)
	if err != nil {
		return nil, err
	}

	runes := make([]rune, 0, length)
	for i := 0; i < len(data); i++ {
		r := charset[data[i]%charsetLen]
		runes = append(runes, r)
	}

	return runes, nil
}