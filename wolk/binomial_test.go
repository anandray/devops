package wolk

/*
import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/stat/distuv"
)

func TestRegularBinomial_Exp(t *testing.T) {
	binomial := NewBinomial(5, 4, 5)
	assert.Equal(t, 0, binomial.Exp(big.NewInt(4), 1).Cmp(big.NewInt(4)))
	assert.Equal(t, 0, binomial.Exp(big.NewInt(4), 2).Cmp(big.NewInt(16)))
	assert.Equal(t, 0, binomial.Exp(big.NewInt(4), 3).Cmp(big.NewInt(64)))
	assert.Equal(t, 0, binomial.Exp(big.NewInt(5), 1).Cmp(big.NewInt(5)))
	assert.Equal(t, 0, binomial.Exp(big.NewInt(5), 2).Cmp(big.NewInt(25)))
	assert.Equal(t, 0, binomial.Exp(big.NewInt(5), 3).Cmp(big.NewInt(125)))
	assert.Equal(t, 0, binomial.Exp(big.NewInt(2), 2).Cmp(big.NewInt(4)))
	assert.Equal(t, 0, binomial.Exp(big.NewInt(2), 4).Cmp(big.NewInt(16)))
}

func TestBinomial_CDF(t *testing.T) {
	binomial := NewBinomial(5, 1, 2)
	b0 := big.NewRat(1, 32)
	b1 := big.NewRat(6, 32)
	b2 := big.NewRat(16, 32)
	b3 := big.NewRat(26, 32)
	b4 := big.NewRat(31, 32)
	b5 := big.NewRat(32, 32)
	assert.Equal(t, 0, binomial.CDF(0).Cmp(b0))
	assert.Equal(t, 0, binomial.CDF(1).Cmp(b1))
	assert.Equal(t, 0, binomial.CDF(2).Cmp(b2))
	assert.Equal(t, 0, binomial.CDF(3).Cmp(b3))
	assert.Equal(t, 0, binomial.CDF(4).Cmp(b4))
	assert.Equal(t, 0, binomial.CDF(5).Cmp(b5))

	binomial2 := NewBinomial(4, 1, 3)
	c0 := big.NewRat(16, 81)
	c1 := big.NewRat(48, 81)
	c2 := big.NewRat(72, 81)
	c3 := big.NewRat(80, 81)
	c4 := big.NewRat(81, 81)
	assert.Equal(t, 0, binomial2.CDF(0).Cmp(c0))
	assert.Equal(t, 0, binomial2.CDF(1).Cmp(c1))
	assert.Equal(t, 0, binomial2.CDF(2).Cmp(c2))
	assert.Equal(t, 0, binomial2.CDF(3).Cmp(c3))
	assert.Equal(t, 0, binomial2.CDF(4).Cmp(c4))
}

func BenchmarkBinomial_CDF(b *testing.B) {
	b.ResetTimer()
	binomial_1 := NewBinomial(1000, 1, 1000000)
	for i := 0; i < b.N; i++ {
		binomial_1.CDF(int64(i))
	}
}

func BenchmarkBinomial_distuv(b *testing.B) {
	b.ResetTimer()
	binomial_2 := &distuv.Binomial{
		N: 1000,
		P: float64(1) / float64(100),
	}
	for i := 0; i < b.N; i++ {
		binomial_2.CDF(float64(i))
	}
}
*/
