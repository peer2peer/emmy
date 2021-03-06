/*
 * Copyright 2017 XLAB d.o.o.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package groups

import (
	"crypto/dsa"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/xlab-si/emmy/crypto/common"
)

// SchnorrGroup is a cyclic group in modular arithmetic. It holds P = Q * R + 1 for some R.
// The actual value R is never used (although a random element from this group could be computed
// by a^R for some random a from Z_p* - this element would have order Q and would be thus from this group),
// the important thing is that Q divides P-1.
type SchnorrGroup struct {
	P *big.Int // modulus of the group
	G *big.Int // generator of subgroup
	Q *big.Int // order of G
}

// NewSchnorrGroup generates random SchnorrGroup with generator G and
// parameters P and Q where P = R * Q + 1 for some R. Order of G is Q.
func NewSchnorrGroup(qBitLength int) (*SchnorrGroup, error) {
	// Using DSA GenerateParameters:
	sizes := dsa.L1024N160
	if qBitLength == 160 {
		sizes = dsa.L1024N160
	} else if qBitLength == 224 {
		sizes = dsa.L2048N224
	} else if qBitLength == 256 {
		sizes = dsa.L2048N256
	} else {
		err := fmt.Errorf("generating Schnorr primes for bit length %d is not supported", qBitLength)
		return nil, err
	}

	params := dsa.Parameters{}
	err := dsa.GenerateParameters(&params, rand.Reader, sizes)
	if err != nil {
		return nil, err
	}

	return &SchnorrGroup{
		P: params.P,
		G: params.G,
		Q: params.Q,
	}, nil

}

func NewSchnorrGroupFromParams(p, g, q *big.Int) *SchnorrGroup {
	return &SchnorrGroup{
		P: p,
		G: g,
		Q: q,
	}
}

// GetRandomElement returns a random element from this group. Note that elements from this group
// are integers smaller than group.P, but not all - only Q of them. GetRandomElement returns
// one (random) of these Q elements.
func (group *SchnorrGroup) GetRandomElement() *big.Int {
	r := common.GetRandomInt(group.Q)
	el := group.Exp(group.G, r)
	return el
}

// Add computes x + y in SchnorrGroup. This means x + y mod group.P.
func (group *SchnorrGroup) Add(x, y *big.Int) *big.Int {
	r := new(big.Int)
	r.Add(x, y)
	r.Mod(r, group.P)
	return r
}

// Mul computes x * y in SchnorrGroup. This means x * y mod group.P.
func (group *SchnorrGroup) Mul(x, y *big.Int) *big.Int {
	r := new(big.Int)
	r.Mul(x, y)
	return r.Mod(r, group.P)
}

// Exp computes base^exponent in SchnorrGroup. This means base^exponent mod group.P.
func (group *SchnorrGroup) Exp(base, exponent *big.Int) *big.Int {
	if exponent.Sign() == -1 { // exponent is negative
		expAbs := new(big.Int).Abs(exponent)
		t := new(big.Int).Exp(base, expAbs, group.P)
		return group.Inv(t)
	}

	return new(big.Int).Exp(base, exponent, group.P)
}

// Inv computes inverse of x in SchnorrGroup. This means xInv such that x * xInv = 1 mod group.P.
func (group *SchnorrGroup) Inv(x *big.Int) *big.Int {
	return new(big.Int).ModInverse(x, group.P)
}

// IsElementInGroup returns true if x is in the group and false otherwise. Note that
// an element x is in Schnorr group when x^group.Q = 1 mod group.P.
func (group *SchnorrGroup) IsElementInGroup(x *big.Int) bool {
	check := group.Exp(x, group.Q) // should be 1
	return check.Cmp(big.NewInt(1)) == 0
}
